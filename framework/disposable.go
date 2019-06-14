package framework

// Disposable - a class/interface would implement this in order to clean up resources
type Disposable interface {
	Dispose()
}
