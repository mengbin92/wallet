package iwallet

// IStorage defines an interface related to data storage.
type IStorage interface {
	// SaveKey saves a key to the storage.
	SaveKey(key string) error
	// LoadKey loads a key from the storage.
	LoadKey() (string, error)
}


