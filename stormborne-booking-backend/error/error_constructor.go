package error

import "net/http"

func NotFoundError(code string, message string, err error) *AppError {
	return NewAppError(http.StatusNotFound, code, message, err)
}

func UnProcessableError(code string, message string, err error) *AppError {
	return NewAppError(http.StatusUnprocessableEntity, code, message, err)
}

func InternalServerError(code string, message string, err error) *AppError {
	return NewAppError(http.StatusInternalServerError, code, message, err)
}

func BadRequestError(code string, message string, err error) *AppError {
	return NewAppError(http.StatusBadRequest, code, message, err)
}

func NewAppError(httpCode int, code string, message string, err error) *AppError {
	return &AppError{httpCode: httpCode, Code: code, Message: message, error: err}
}

func InvalidCredentialsError(code string, message string, err error) *AppError {
	return NewAppError(http.StatusUnauthorized, code, message, err)
}

func ErrPickEitherIDOrURL() *AppError {
	return NewAppError(http.StatusBadRequest, "ERR_PICK_EITHER_ID_OR_URL", "provide exactly one of predefined_avatar_id or avatar_url", nil)
}

func ErrInvalidAvatarType() *AppError {
	return NewAppError(http.StatusBadRequest, "ERR_INVALID_AVATAR_TYPE", "invalid avatar_type", nil)
}

func ErrInvalidAvatarURL() *AppError {
	return NewAppError(http.StatusBadRequest, "ERR_INVALID_AVATAR_URL", "invalid avatar_url", nil)
}

func ErrAvatarNotFound() *AppError {
	return NewAppError(http.StatusNotFound, "ERR_AVATAR_NOT_FOUND", "predefined avatar not found", nil)
}

func ErrImageTooLarge() *AppError {
	return NewAppError(http.StatusBadRequest, "ERR_IMAGE_TOO_LARGE", "image size exceeds the 5MB limit", nil)
}

func ErrImageUploadFailed(err error) *AppError {
	return NewAppError(http.StatusInternalServerError, "ERR_IMAGE_UPLOAD_FAILED", "failed to upload image to storage service", err)
}

func ErrSirvTokenFailed(err error) *AppError {
	return NewAppError(http.StatusInternalServerError, "ERR_SIRV_TOKEN_FAILED", "failed to authenticate with image service", err)
}
