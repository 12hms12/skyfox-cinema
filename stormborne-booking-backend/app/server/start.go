package server

import (
	"fmt"
	"net/http"
	"time"

	"skyfox/bookings/constants"
	"skyfox/bookings/controller"
	"skyfox/bookings/database/connection"
	database "skyfox/bookings/database/seed"
	"skyfox/bookings/repository"
	"skyfox/bookings/service"
	"skyfox/common/logger"
	"skyfox/common/middleware/cors"
	"skyfox/common/middleware/security"
	"skyfox/common/middleware/validator"
	appConf "skyfox/config"
	movieservice "skyfox/movieservice/movie_gateway"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	_ "skyfox/docs" //indirect

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Init(cfg appConf.AppConfig) error {

	logger.InitAppLogger(cfg.Logger)

	handler := connection.NewDBHandler(cfg.Database)
	db := handler.Instance()
	jwtManager := security.NewJwtManager(cfg.Token)

	movieGateway := movieservice.NewMovieGateway(cfg.MovieGateway)

	bookingRepository := repository.NewBookingRepository(db)
	onlineCustomerRepository := repository.NewCustomerRepository(db)
	avatarRepository := repository.NewAvatarRepository(db)
	profilePictureRepository := repository.NewProfilePictureRepository(db)
	otpRepository := repository.NewOTPRepository(db)
	seatRepo := repository.NewSeatRepository(db)
	showRepository := repository.NewShowRepository(db, movieGateway, seatRepo)
	customerBookingRepo := repository.NewCustomerBookingRepository(db)
	bookingViewRepository := repository.NewBookingViewRepository(db)

	database.SeedAll(db.GormDB(),onlineCustomerRepository,seatRepo,showRepository,database.SeedingConfig{})

	// instantiate all services
	showService := service.NewShowService(showRepository, movieGateway)
	adminScheduleService := service.NewAdminScheduleService(showRepository, movieGateway)
	revenueService := service.NewRevenueService(bookingRepository, showRepository)
	avatarService := service.NewAvatarService(avatarRepository, profilePictureRepository)
	profilePictureService := service.NewProfileImageService(avatarRepository, profilePictureRepository)
	movieService := service.NewMovieService(movieGateway)
	smsOtpService := service.NewSMSOtpService(otpRepository)
	emailOtpService := service.NewEmailOtpService(otpRepository)
	customerService := service.NewCustomerService(onlineCustomerRepository, jwtManager, avatarService, &cfg.Server, otpRepository, emailOtpService)
	seatService := service.NewSeatService(seatRepo)
	captchaService := service.NewCaptchaService()
	customerBookingService := service.NewCustomerBookingService(customerBookingRepo, seatRepo)
	bookingViewService := service.NewBookingViewService(bookingViewRepository, movieGateway)
	ownerAdminService := service.NewOwnerAdminService(onlineCustomerRepository)

	customerBookingController := controller.NewCustomerBookingController(customerBookingService)
	showController := controller.NewShowController(showService)
	adminScheduleController := controller.NewAdminScheduleController(adminScheduleService)
	revenueController := controller.NewRevenueController(revenueService)
	onlineCustomerController := controller.NewOnlineCustomerController(customerService, captchaService)
	movieController := controller.NewMovieController(movieService)
	avatarController := controller.NewAvatarController(avatarService)
	otpController := controller.NewOTPController(smsOtpService)
	emailOtpController := controller.NewEmailOTPController(emailOtpService)
	seatController := controller.NewSeatController(seatService)
	captchaController := controller.NewCaptchaController(captchaService)
	bookingViewController := controller.NewBookingViewController(bookingViewService)
	profileImageController := controller.NewProfileImageController(profilePictureService)
	ownerAdminController := controller.NewOwnerAdminController(ownerAdminService)

	router := setupApp(cfg)

	adminRouter := routerGroupWithAllowedRoles(router, jwtManager, security.ADMIN, security.OWNER)
	customerRouter := routerGroupWithAllowedRoles(router, jwtManager, security.USER)
	commonRouter := routerGroupWithAllowedRoles(router, jwtManager, security.USER, security.ADMIN, security.OWNER)
	ownerRouter := routerGroupWithAllowedRoles(router, jwtManager, security.OWNER)
	noAuthRouter := routerGroupWithNoAuth(router)

	revenue := adminRouter.Group(constants.RevenueEndPoint)
	{
		revenue.GET("", revenueController.GetRevenue)
	}

	show := noAuthRouter.Group(constants.ShowEndPoint)
	{
		show.GET("", showController.Shows)
	}

	adminSchedule := adminRouter.Group(constants.AdminWeeklyScheduleEndPoint)
	{
		adminSchedule.GET("", adminScheduleController.GetWeeklySchedule)
	}

	adminRouter.GET(constants.AdminSlotsEndPoint, adminScheduleController.GetSlots)
	adminRouter.GET(constants.AdminScreensEndPoint, adminScheduleController.GetScreens)
	adminRouter.POST(constants.AdminScreensEndPoint, adminScheduleController.AddScreen)

	adminShows := adminRouter.Group(constants.AdminShowsEndpoint)
	{
		adminShows.POST("", adminScheduleController.ScheduleShow)
		adminShows.DELETE("/:showId", adminScheduleController.DeleteShow)
	}

	onlineCustomer := noAuthRouter.Group(constants.ForgotPasswordEndPoint)
	{
		onlineCustomer.POST("", onlineCustomerController.ForgotPassword)
	}

	onlineCustomerReset := noAuthRouter.Group(constants.ResetPasswordEndPoint)
	{
		onlineCustomerReset.POST("", onlineCustomerController.ResetPassword)
	}

	onlineCustomerResetPage := noAuthRouter.Group(constants.ResetPasswordEndPoint)
	{
		onlineCustomerResetPage.GET("", onlineCustomerController.ResetPasswordPage)
	}

	adminRouter.GET(constants.MoviesEndPoint, movieController.Movies)
	adminRouter.GET(constants.AdminMoviesEndPoint, movieController.Movies)
	adminRouter.GET(constants.AdminMoviesEndPoint+"/:movieId", movieController.MovieByID)
	adminRouter.GET(constants.AdminViewBookingsEndPoint, bookingViewController.GetBookings)
	adminRouter.PATCH(constants.CheckedInBookingEndpoint, customerBookingController.CheckedIn)

	ownerAdmins := ownerRouter.Group(constants.OwnerAdminsEndPoint)
	{
		ownerAdmins.GET("", ownerAdminController.ListAdmins)
		ownerAdmins.POST("", ownerAdminController.AddAdmin)
		ownerAdmins.DELETE("/:adminId", ownerAdminController.RemoveAdmin)
	}

	avatar := customerRouter.Group("/avatar")
	{
		avatar.GET("", avatarController.GetMyAvatar)
		avatar.PUT("", avatarController.UpdateMyAvatar)
	}
	noAuthRouter.GET("/avatar/predefined", avatarController.ListPredefinedAvatars)

	profileImage := customerRouter.Group(constants.ProfileImageEndpoint)
	{
		profileImage.POST("", profileImageController.UploadProfileImage)
		profileImage.GET("", profileImageController.GetProfileImage)
	}

	noAuthRouter.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	noAuthRouter.POST(constants.OnlineCustomerSignUp, onlineCustomerController.SignupController)
	noAuthRouter.POST(constants.LoginEndPoint, onlineCustomerController.LoginController)
	customerRouter.GET(constants.CustomerShowsEndpoint, showController.Shows)
	customerRouter.GET(constants.CustomerBooking, bookingViewController.GetAllBookingsCustomer)
	commonRouter.GET(constants.UserProfileEndpoint, onlineCustomerController.GetProfileController)
	commonRouter.PUT(constants.UserProfileEndpoint, onlineCustomerController.UpdateProfileController)
	noAuthRouter.GET("/", showController.Shows)

	noAuthRouter.GET(constants.VerifyResetTokenEndPoint, onlineCustomerController.VerifyResetToken)

	noAuthRouter.POST(constants.OTPSMSRequestEndpoint, otpController.RequestSMSOTP)
	noAuthRouter.POST(constants.OTPSMSVerifyEndpoint, otpController.VerifySMSOTP)
	noAuthRouter.POST(constants.OTPSMSResendEndpoint, otpController.ResendSMSOTP)
	noAuthRouter.GET(constants.OTPListAllEndpoint, otpController.ListAllOTPs)

	noAuthRouter.POST(constants.OTPEmailRequestEndpoint, emailOtpController.RequestEmailOTP)
	noAuthRouter.POST(constants.OTPEmailVerifyEndpoint, emailOtpController.VerifyEmailOTP)
	noAuthRouter.POST(constants.OTPEmailResendEndpoint, emailOtpController.ResendEmailOTP)

	noAuthRouter.GET(constants.SeatSelectionEndpoint, seatController.GetSeatStatus)

	noAuthRouter.GET(constants.ScheduledMoviesEndpoint, showController.ScheduledMovies)
	noAuthRouter.GET(constants.MovieShowtimesEndpoint, showController.MovieShowtimes)
	noAuthRouter.GET(constants.MovieAvailableDatesEndpoint, showController.MovieAvailableDates)

	noAuthRouter.GET("/captcha/generate", captchaController.GenerateCaptcha)
	noAuthRouter.POST("/captcha/verify", captchaController.VerifyCaptcha)
	noAuthRouter.POST(constants.PaymentEndpoint, controller.HandlePayment)

	customerBookings := commonRouter.Group(constants.CustomerBookingsEndpoint)
	{
		customerBookings.POST("", customerBookingController.CreateBooking)
	}
	commonRouter.POST(constants.CustomerBookingPaymentEndpoint, customerBookingController.ProcessPayment)
	commonRouter.GET(constants.CustomerBookingDetailEndpoint, customerBookingController.GetBookingDetails)

	err := start(router, cfg.Server)
	if err != nil {
		return err
	}
	return nil
}

func start(r *gin.Engine, cfg appConf.ServerConfig) error {
	s := &http.Server{
		Addr:         port(cfg),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
	}
	err := s.ListenAndServe()
	if err != nil {
		return fmt.Errorf("unable to start gin server. error: %w", err)
	}
	return nil
}

func setupApp(cfg appConf.AppConfig) *gin.Engine {
	gin.SetMode(cfg.Server.GineMode)
	engine := gin.New()
	binding.Validator = new(validator.DtoValidator)
	return setupMiddleware(engine)
}

func setupMiddleware(engine *gin.Engine) *gin.Engine {
	engine.Use(cors.SetupCORS())
	engine.Use(ginzap.Ginzap(logger.GetLogger(), time.RFC3339, true))
	engine.Use(ginzap.RecoveryWithZap(logger.GetLogger(), true))
	return engine
}

func port(c appConf.ServerConfig) string {
	return fmt.Sprintf(":%d", c.Port)
}

func routerGroupWithAllowedRoles(engine *gin.Engine, jwtManager *security.JwtManager, allowedRoles ...security.Role) *gin.RouterGroup {
	authRouter := engine.Group("")
	authRouter.Use(security.JWTAuth(jwtManager, allowedRoles...))
	return authRouter
}

func routerGroupWithNoAuth(engine *gin.Engine) *gin.RouterGroup {
	noAuthRouter := engine.Group("")
	return noAuthRouter
}