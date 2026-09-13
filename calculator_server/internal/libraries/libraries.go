package libraries

import "github.com/ebitengine/purego"

type CLibrary struct {
	Add     func(a, b int64) int64
	handles []uintptr
}

func (l *CLibrary) Close() {
	if l == nil {
		return
	}

	for _, h := range l.handles {
		if h != 0 {
			purego.Dlclose(h)
		}
	}
	l.handles = nil
}

type RustLibrary struct {
	Sub     func(a, b int64) int64
	handles []uintptr
}

func (l *RustLibrary) Close() {
	if l == nil {
		return
	}

	for _, h := range l.handles {
		if h != 0 {
			purego.Dlclose(h)
		}
	}
	l.handles = nil
}

func Load(cLibPath, rustLibPath string) (*CLibrary, *RustLibrary, error) {
	var loadErr error

	cLibrary := &CLibrary{handles: make([]uintptr, 0)}
	defer func() {
		if loadErr != nil {
			cLibrary.Close()
		}
	}()

	cHandle, err := purego.Dlopen(cLibPath, purego.RTLD_NOW)
	if err != nil {
		loadErr = err
		return nil, nil, err
	}
	cLibrary.handles = append(cLibrary.handles, cHandle)

	cAddPtr, err := purego.Dlsym(cHandle, "add")
	if err != nil {
		loadErr = err
		return nil, nil, err
	}
	purego.RegisterFunc(&cLibrary.Add, cAddPtr)

	rustLibrary := &RustLibrary{handles: make([]uintptr, 0)}
	defer func() {
		if loadErr != nil {
			rustLibrary.Close()
		}
	}()

	rustHandle, err := purego.Dlopen(rustLibPath, purego.RTLD_NOW)
	if err != nil {
		loadErr = err
		return nil, nil, err
	}
	rustLibrary.handles = append(rustLibrary.handles, rustHandle)

	rustSubPtr, err := purego.Dlsym(rustHandle, "sub")
	if err != nil {
		loadErr = err
		return nil, nil, err
	}
	purego.RegisterFunc(&rustLibrary.Sub, rustSubPtr)

	return cLibrary, rustLibrary, nil
}
