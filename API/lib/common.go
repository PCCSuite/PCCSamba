package lib

type PasswordMode int

const (
	PasswordModeDynamic         PasswordMode = 0
	PasswordModeStaticPlain     PasswordMode = 1
	PasswordModeStaticEncrypted PasswordMode = 2
	PasswordModeStaticUnstored  PasswordMode = 3
)
