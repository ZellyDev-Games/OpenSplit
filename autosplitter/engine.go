package autosplitter

import lua "github.com/yuin/gopher-lua"

type Engine struct {
	L               *lua.LState
	risingSigned    *lua.LFunction
	fallingSigned   *lua.LFunction
	risingUnsigned  *lua.LFunction
	fallingUnsigned *lua.LFunction
	edge            *lua.LFunction
}

func NewEngine() *Engine {
	L := lua.NewState()
	return &Engine{L: L}
}

func (e *Engine) Close() {
	if e.L != nil {
		e.L.Close()
	}
}

func (e *Engine) LoadFile(path string) error {
	if err := e.L.DoFile(path); err != nil {
		return err
	}
	e.cacheCallbacks()
	return nil
}

func (e *Engine) RisingSigned(id string, old, new int64) {
	if e.risingSigned == nil {
		return
	}
	_ = e.L.CallByParam(lua.P{
		Fn:      e.risingSigned,
		NRet:    0,
		Protect: true,
	}, lua.LString(id), lua.LNumber(old), lua.LNumber(new))
}

func (e *Engine) FallingSigned(id string, old, new int64) {
	if e.fallingSigned == nil {
		return
	}
	_ = e.L.CallByParam(lua.P{
		Fn:      e.fallingSigned,
		NRet:    0,
		Protect: true,
	}, lua.LString(id), lua.LNumber(old), lua.LNumber(new))
}

func (e *Engine) RisingUnsigned(id string, old, new uint64) {
	if e.risingUnsigned == nil {
		return
	}
	_ = e.L.CallByParam(lua.P{
		Fn:      e.risingUnsigned,
		NRet:    0,
		Protect: true,
	}, lua.LString(id), lua.LNumber(old), lua.LNumber(new))
}

func (e *Engine) FallingUnsigned(id string, old, new uint64) {
	if e.fallingUnsigned == nil {
		return
	}
	_ = e.L.CallByParam(lua.P{
		Fn:      e.fallingUnsigned,
		NRet:    0,
		Protect: true,
	}, lua.LString(id), lua.LNumber(old), lua.LNumber(new))
}

func (e *Engine) Edge(id string, new bool) {
	if e.edge == nil {
		return
	}
	var b lua.LValue = lua.LFalse
	if new {
		b = lua.LTrue
	}
	_ = e.L.CallByParam(lua.P{
		Fn:      e.edge,
		NRet:    0,
		Protect: true,
	}, lua.LString(id), b)
}

func (e *Engine) cacheCallbacks() {
	e.risingSigned = asFunc(e.L.GetGlobal("risingSigned"))
	e.fallingSigned = asFunc(e.L.GetGlobal("fallingSigned"))
	e.risingUnsigned = asFunc(e.L.GetGlobal("risingUnsigned"))
	e.fallingUnsigned = asFunc(e.L.GetGlobal("fallingUnsigned"))
	e.edge = asFunc(e.L.GetGlobal("edge"))
}

func asFunc(v lua.LValue) *lua.LFunction {
	if v == lua.LNil {
		return nil
	}
	if f, ok := v.(*lua.LFunction); ok {
		return f
	}
	return nil
}
