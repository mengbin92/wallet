package iwallet

type IKey2Address interface {
	// GenP2PKAddress Generates the BTC Pay-to-Pubkey address
	GenP2PKAddress(network string) (string, error)
	// GenP2PKHAddress Generates the BTC Pay-to-Pubkey-Hash
	GenP2PKHAddress(network string) (string, error)
	// GenBech32Address Generates the BTC SegWit address
	GenBech32Address(network string) (string, error)

	// ExportPrivateKey Exports the private key
	ExportPrivateKey(pwd string) (string, error)
	// LoadPrivateKey Loads the private key
	LoadPrivateKey(encryptStr, pwd string) error
}
