package pkg

import "github.com/ebitengine/purego"

type CLib struct {
	add     func(a, b int64) int64
	handles []uintptr
}

func (c *CLib) Add(a, b int64) int64 {
	if c.add != nil {
		return c.add(a, b)
	}
	return 0
}
func (c *CLib) Close() {
	if c == nil {
		return
	}

	for _, h := range c.handles {
		if h != 0 {
			purego.Dlclose(h)
		}
	}
	c.handles = nil
}

type RustLib struct {
	sub     func(a, b int64) int64
	handles []uintptr
}

func (r *RustLib) Sub(a, b int64) int64 {
	if r.sub != nil {
		return r.sub(a, b)
	}
	return 0
}
func (r *RustLib) Close() {
	if r == nil {
		return
	}

	for _, h := range r.handles {
		if h != 0 {
			purego.Dlclose(h)
		}
	}
	r.handles = nil
}

func LoadLibraries(cLibPath, rustLibPath string) (*CLib, *RustLib, error) {
	var loadErr error

	cLib := &CLib{handles: make([]uintptr, 0)}
	defer func() {
		if loadErr != nil {
			cLib.Close()
		}
	}()

	cHandle, err := purego.Dlopen(cLibPath, purego.RTLD_NOW)
	if err != nil {
		loadErr = err
		return nil, nil, err
	}
	cLib.handles = append(cLib.handles, cHandle)

	cAddPtr, err := purego.Dlsym(cHandle, "add")
	if err != nil {
		loadErr = err
		return nil, nil, err
	}
	purego.RegisterFunc(&cLib.add, cAddPtr)

	rustLib := &RustLib{handles: make([]uintptr, 0)}
	defer func() {
		if loadErr != nil {
			rustLib.Close()
		}
	}()

	rustHandle, err := purego.Dlopen(rustLibPath, purego.RTLD_NOW)
	if err != nil {
		loadErr = err
		return nil, nil, err
	}
	rustLib.handles = append(rustLib.handles, rustHandle)

	rustSubPtr, err := purego.Dlsym(rustHandle, "sub")
	if err != nil {
		loadErr = err
		return nil, nil, err
	}
	purego.RegisterFunc(&rustLib.sub, rustSubPtr)

	return cLib, rustLib, nil
}
