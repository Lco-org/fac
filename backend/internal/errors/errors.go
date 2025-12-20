// Package errors provides structured error handling for CSO-ERP
package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application error with code and context
type AppError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	MessageAr  string                 `json:"message_ar,omitempty"`
	HTTPStatus int                    `json:"-"`
	Err        error                  `json:"-"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithError wraps an underlying error
func (e *AppError) WithError(err error) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		MessageAr:  e.MessageAr,
		HTTPStatus: e.HTTPStatus,
		Err:        err,
		Details:    e.Details,
	}
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		MessageAr:  e.MessageAr,
		HTTPStatus: e.HTTPStatus,
		Err:        e.Err,
		Details:    details,
	}
}

// ToResponse converts error to API response format
func (e *AppError) ToResponse() map[string]interface{} {
	resp := map[string]interface{}{
		"error":      e.Message,
		"error_code": e.Code,
	}
	if e.MessageAr != "" {
		resp["error_ar"] = e.MessageAr
	}
	if len(e.Details) > 0 {
		resp["details"] = e.Details
	}
	return resp
}

// ============================================================================
// ERROR CODES - Authentication
// ============================================================================

var (
	// ErrInvalidCredentials - invalid username or password
	ErrInvalidCredentials = &AppError{
		Code:       "AUTH_001",
		Message:    "Invalid username or password",
		MessageAr:  "اسم المستخدم أو كلمة المرور غير صحيحة",
		HTTPStatus: http.StatusUnauthorized,
	}

	// ErrUserNotActive - user account is not active
	ErrUserNotActive = &AppError{
		Code:       "AUTH_002",
		Message:    "User account is not active",
		MessageAr:  "حساب المستخدم غير نشط",
		HTTPStatus: http.StatusForbidden,
	}

	// ErrInvalidToken - JWT token is invalid or expired
	ErrInvalidToken = &AppError{
		Code:       "AUTH_003",
		Message:    "Invalid or expired session",
		MessageAr:  "الجلسة غير صالحة أو منتهية الصلاحية",
		HTTPStatus: http.StatusUnauthorized,
	}

	// ErrSessionInvalidated - session was revoked
	ErrSessionInvalidated = &AppError{
		Code:       "AUTH_004",
		Message:    "Session has been invalidated. Please login again.",
		MessageAr:  "تم إبطال الجلسة. يرجى تسجيل الدخول مرة أخرى",
		HTTPStatus: http.StatusUnauthorized,
	}

	// Err2FARequired - 2FA verification is required
	Err2FARequired = &AppError{
		Code:       "AUTH_005",
		Message:    "Two-factor authentication required",
		MessageAr:  "التحقق الثنائي مطلوب",
		HTTPStatus: http.StatusAccepted,
	}

	// ErrInvalid2FACode - invalid 2FA verification code
	ErrInvalid2FACode = &AppError{
		Code:       "AUTH_006",
		Message:    "Invalid verification code",
		MessageAr:  "رمز التحقق غير صحيح",
		HTTPStatus: http.StatusUnauthorized,
	}

	// Err2FACodeExpired - 2FA code has expired
	Err2FACodeExpired = &AppError{
		Code:       "AUTH_007",
		Message:    "Verification code has expired",
		MessageAr:  "انتهت صلاحية رمز التحقق",
		HTTPStatus: http.StatusUnauthorized,
	}

	// ErrTooManyAttempts - too many failed attempts
	ErrTooManyAttempts = &AppError{
		Code:       "AUTH_008",
		Message:    "Too many failed attempts. Please try again later.",
		MessageAr:  "محاولات فاشلة كثيرة. يرجى المحاولة لاحقاً",
		HTTPStatus: http.StatusTooManyRequests,
	}

	// ErrWeakPassword - password is too weak
	ErrWeakPassword = &AppError{
		Code:       "AUTH_009",
		Message:    "Password must be at least 8 characters",
		MessageAr:  "يجب أن تكون كلمة المرور 8 أحرف على الأقل",
		HTTPStatus: http.StatusBadRequest,
	}
)

// ============================================================================
// ERROR CODES - Authorization
// ============================================================================

var (
	// ErrForbidden - user does not have permission
	ErrForbidden = &AppError{
		Code:       "AUTHZ_001",
		Message:    "You do not have permission to perform this action",
		MessageAr:  "ليس لديك صلاحية لتنفيذ هذا الإجراء",
		HTTPStatus: http.StatusForbidden,
	}

	// ErrInsufficientRole - user role is insufficient
	ErrInsufficientRole = &AppError{
		Code:       "AUTHZ_002",
		Message:    "Insufficient role permissions",
		MessageAr:  "صلاحيات الدور غير كافية",
		HTTPStatus: http.StatusForbidden,
	}
)

// ============================================================================
// ERROR CODES - Validation
// ============================================================================

var (
	// ErrValidationFailed - request validation failed
	ErrValidationFailed = &AppError{
		Code:       "VAL_001",
		Message:    "Validation failed",
		MessageAr:  "فشل التحقق من البيانات",
		HTTPStatus: http.StatusBadRequest,
	}

	// ErrInvalidEmail - invalid email format
	ErrInvalidEmail = &AppError{
		Code:       "VAL_002",
		Message:    "Invalid email format",
		MessageAr:  "صيغة البريد الإلكتروني غير صالحة",
		HTTPStatus: http.StatusBadRequest,
	}

	// ErrInvalidEmailDomain - email domain not allowed
	ErrInvalidEmailDomain = &AppError{
		Code:       "VAL_003",
		Message:    "Email domain not allowed",
		MessageAr:  "نطاق البريد الإلكتروني غير مسموح به",
		HTTPStatus: http.StatusBadRequest,
	}

	// ErrInvalidUUID - invalid UUID format
	ErrInvalidUUID = &AppError{
		Code:       "VAL_004",
		Message:    "Invalid ID format",
		MessageAr:  "صيغة المعرف غير صالحة",
		HTTPStatus: http.StatusBadRequest,
	}

	// ErrRequiredField - required field is missing
	ErrRequiredField = &AppError{
		Code:       "VAL_005",
		Message:    "Required field is missing",
		MessageAr:  "حقل مطلوب مفقود",
		HTTPStatus: http.StatusBadRequest,
	}
)

// ============================================================================
// ERROR CODES - Resources
// ============================================================================

var (
	// ErrNotFound - resource not found
	ErrNotFound = &AppError{
		Code:       "RES_001",
		Message:    "Resource not found",
		MessageAr:  "المورد غير موجود",
		HTTPStatus: http.StatusNotFound,
	}

	// ErrUserNotFound - user not found
	ErrUserNotFound = &AppError{
		Code:       "RES_002",
		Message:    "User not found",
		MessageAr:  "المستخدم غير موجود",
		HTTPStatus: http.StatusNotFound,
	}

	// ErrTransactionNotFound - transaction not found
	ErrTransactionNotFound = &AppError{
		Code:       "RES_003",
		Message:    "Transaction not found",
		MessageAr:  "المعاملة المالية غير موجودة",
		HTTPStatus: http.StatusNotFound,
	}

	// ErrContainerNotFound - container not found
	ErrContainerNotFound = &AppError{
		Code:       "RES_004",
		Message:    "Financial container not found",
		MessageAr:  "الحاوية المالية غير موجودة",
		HTTPStatus: http.StatusNotFound,
	}

	// ErrProjectNotFound - project not found
	ErrProjectNotFound = &AppError{
		Code:       "RES_005",
		Message:    "Project not found",
		MessageAr:  "المشروع غير موجود",
		HTTPStatus: http.StatusNotFound,
	}

	// ErrOrganizationNotFound - organization not found
	ErrOrganizationNotFound = &AppError{
		Code:       "RES_006",
		Message:    "Organization not found",
		MessageAr:  "المنظمة غير موجودة",
		HTTPStatus: http.StatusNotFound,
	}

	// ErrDuplicateEntry - duplicate entry
	ErrDuplicateEntry = &AppError{
		Code:       "RES_007",
		Message:    "Entry already exists",
		MessageAr:  "السجل موجود مسبقاً",
		HTTPStatus: http.StatusConflict,
	}
)

// ============================================================================
// ERROR CODES - Transactions
// ============================================================================

var (
	// ErrInsufficientBalance - insufficient balance for transaction
	ErrInsufficientBalance = &AppError{
		Code:       "TXN_001",
		Message:    "Insufficient balance",
		MessageAr:  "الرصيد غير كافٍ",
		HTTPStatus: http.StatusBadRequest,
	}

	// ErrTransactionAlreadyReversed - transaction already reversed
	ErrTransactionAlreadyReversed = &AppError{
		Code:       "TXN_002",
		Message:    "Transaction has already been reversed",
		MessageAr:  "تم عكس المعاملة مسبقاً",
		HTTPStatus: http.StatusBadRequest,
	}

	// ErrInvalidTransactionType - invalid transaction type
	ErrInvalidTransactionType = &AppError{
		Code:       "TXN_003",
		Message:    "Invalid transaction type",
		MessageAr:  "نوع المعاملة غير صالح",
		HTTPStatus: http.StatusBadRequest,
	}

	// ErrInvalidAmount - invalid amount
	ErrInvalidAmount = &AppError{
		Code:       "TXN_004",
		Message:    "Amount must be greater than zero",
		MessageAr:  "يجب أن يكون المبلغ أكبر من صفر",
		HTTPStatus: http.StatusBadRequest,
	}
)

// ============================================================================
// ERROR CODES - Multi-tenant
// ============================================================================

var (
	// ErrTenantNotFound - tenant not found
	ErrTenantNotFound = &AppError{
		Code:       "TEN_001",
		Message:    "Tenant not found",
		MessageAr:  "المستأجر غير موجود",
		HTTPStatus: http.StatusNotFound,
	}

	// ErrTenantSuspended - tenant is suspended
	ErrTenantSuspended = &AppError{
		Code:       "TEN_002",
		Message:    "Organization is suspended",
		MessageAr:  "المنظمة معلقة",
		HTTPStatus: http.StatusForbidden,
	}

	// ErrDomainNotVerified - domain not verified
	ErrDomainNotVerified = &AppError{
		Code:       "TEN_003",
		Message:    "Domain not verified",
		MessageAr:  "النطاق غير مُتحقق منه",
		HTTPStatus: http.StatusForbidden,
	}

	// ErrDomainAlreadyExists - domain already exists
	ErrDomainAlreadyExists = &AppError{
		Code:       "TEN_004",
		Message:    "Domain already registered",
		MessageAr:  "النطاق مسجل مسبقاً",
		HTTPStatus: http.StatusConflict,
	}
)

// ============================================================================
// ERROR CODES - System
// ============================================================================

var (
	// ErrInternal - internal server error
	ErrInternal = &AppError{
		Code:       "SYS_001",
		Message:    "An internal error occurred",
		MessageAr:  "حدث خطأ داخلي",
		HTTPStatus: http.StatusInternalServerError,
	}

	// ErrDatabaseError - database error
	ErrDatabaseError = &AppError{
		Code:       "SYS_002",
		Message:    "Database error occurred",
		MessageAr:  "حدث خطأ في قاعدة البيانات",
		HTTPStatus: http.StatusInternalServerError,
	}

	// ErrEmailSendFailed - email sending failed
	ErrEmailSendFailed = &AppError{
		Code:       "SYS_003",
		Message:    "Failed to send email",
		MessageAr:  "فشل إرسال البريد الإلكتروني",
		HTTPStatus: http.StatusInternalServerError,
	}

	// ErrServiceUnavailable - service temporarily unavailable
	ErrServiceUnavailable = &AppError{
		Code:       "SYS_004",
		Message:    "Service temporarily unavailable",
		MessageAr:  "الخدمة غير متاحة مؤقتاً",
		HTTPStatus: http.StatusServiceUnavailable,
	}
)

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// New creates a new AppError
func New(code, message string, status int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
	}
}

// Wrap wraps an error with an AppError
func Wrap(appErr *AppError, err error) *AppError {
	return appErr.WithError(err)
}

// IsAppError checks if error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError extracts AppError from error
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return ErrInternal.WithError(err)
}

// GetHTTPStatus returns HTTP status code for error
func GetHTTPStatus(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}
