package autosplitter

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zellydev-games/opensplit/dispatcher"
)

type Engine struct {
	L               *lua.LState
	dispatcher      *dispatcher.Service
	risingSigned    *lua.LFunction
	fallingSigned   *lua.LFunction
	risingUnsigned  *lua.LFunction
	fallingUnsigned *lua.LFunction
	edge            *lua.LFunction
}

func NewEngine(d *dispatcher.Service) *Engine {
	L := lua.NewState()
	e := &Engine{
		L:          L,
		dispatcher: d,
	}

	e.L.SetGlobal("split", e.L.NewFunction(func(L *lua.LState) int {
		d.Dispatch(dispatcher.SPLIT, nil)
		return 0
	}))

	e.L.SetGlobal("reset", e.L.NewFunction(func(L *lua.LState) int {
		d.Dispatch(dispatcher.RESET, nil)
		return 0
	}))

	e.L.SetGlobal("pause", e.L.NewFunction(func(L *lua.LState) int {
		d.Dispatch(dispatcher.PAUSE, nil)
		return 0
	}))

	return e
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
