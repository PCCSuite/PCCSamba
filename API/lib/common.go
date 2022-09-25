package lib

type PasswordMode int

const (
	PasswordModeDynamic         PasswordMode = 1
	PasswordModeStaticPlain     PasswordMode = 2
	PasswordModeStaticEncrypted PasswordMode = 3
	PasswordModeStaticUnstored  PasswordMode = 4
)
