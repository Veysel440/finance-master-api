package errs

var (
	AccountLocked   = E("account_locked", 423, "account temporarily locked")
	SessionMismatch = E("session_mismatch", 401, "refresh not valid for this device/ip")
)
