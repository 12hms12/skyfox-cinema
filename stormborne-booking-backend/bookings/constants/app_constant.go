package constants

const (
	LoginEndPoint                  = "/login"
	ForgotPasswordEndPoint         = "/forgot-password"
	ResetPasswordEndPoint          = "/reset-password"
	VerifyResetTokenEndPoint       = "/verify-reset-token"
	OnlineCustomerSignUp           = "/customer/signup"
	AdminShowsEndpoint             = "/admin/shows"
	AdminMoviesEndPoint            = "/admin/movies"
	AdminWeeklyScheduleEndPoint    = "/admin/schedule/week"
	AdminSlotsEndPoint             = "/admin/slots"
	AdminViewBookingsEndPoint      = "/admin/view-bookings"
	AdminScreensEndPoint           = "/admin/screens"
	OwnerAdminsEndPoint            = "/owner/admins"
	RevenueEndPoint                = "/revenue"
	CustomerShowsEndpoint          = "/customer/shows"
	UserProfileEndpoint            = "/profile"
	BookingEndPoint                = "/bookings"
	ShowEndPoint                   = "/shows"
	MoviesEndPoint                 = "/movies"
	OTPSMSRequestEndpoint          = "/otp/sms/request"
	OTPSMSVerifyEndpoint           = "/otp/sms/verify"
	OTPSMSResendEndpoint           = "/otp/sms/resend"
	OTPListAllEndpoint             = "/otp/all"
	OTPEmailRequestEndpoint        = "/otp/email/request"
	OTPEmailVerifyEndpoint         = "/otp/email/verify"
	OTPEmailResendEndpoint         = "/otp/email/resend"
	SeatSelectionEndpoint          = "/shows/:showId/seats"
	CustomerBookingsEndpoint       = "/customer/bookings"
	CustomerBookingPaymentEndpoint = "/customer/bookings/:bookingId/payment"
	CustomerBookingDetailEndpoint  = "/customer/bookings/:bookingId"
	CheckedInBookingEndpoint       = "/admin/booking-checked-in/:bookingId"
	PaymentEndpoint      = "/payment"
	ProfileImageEndpoint = "/profile/profile-image"

	ScheduledMoviesEndpoint     = "/movies/scheduled"
	MovieShowtimesEndpoint      = "/movies/:movieId/showtimes"
	MovieAvailableDatesEndpoint = "/movies/:movieId/available-dates"
	CustomerBooking = "bookings/:onlineCustomerId"
)

const (
	MaxProfileImageSizeBytes = 5 * 1024 * 1024
)

const (
	TOTAL_NO_OF_SEATS           = 100
	MAX_NO_OF_SEATS_PER_BOOKING = 15
)
const (
	Pg_duplicate_error = "23505"
	StockImage         = "https://t4.ftcdn.net/jpg/06/57/37/01/360_F_657370150_pdNeG5pjI976ZasVbKN9VqH1rfoykdYU.jpg"
)

const (
	CARD_LENGTH     = 16
	CVV_LENGTH      = 3
	MIN_NAME_LENGTH = 3
)
