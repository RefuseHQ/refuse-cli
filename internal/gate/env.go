package gate

import "os"

// envIfSet wraps os.LookupEnv so we can stub it in tests if needed.
func envIfSet(name string) (string, bool) { return os.LookupEnv(name) }
