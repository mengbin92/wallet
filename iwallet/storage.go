package iwallet

// IStorage defines an interface related to data storage.
type IKeyStorage interface {
	// SaveKey saves a key to the storage.
	SaveKey(key string) error
	// ListKeys lists all keys in the storage.
	ListKeys() ([]string, error)
}


