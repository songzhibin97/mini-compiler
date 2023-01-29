package compiler

type SymbolScope string

const (
	GlobalScope   SymbolScope = "GlobalScope"
	LocalScope    SymbolScope = "LocalScope"
	BuiltinScope  SymbolScope = "BuiltinScope"
	ContextScope  SymbolScope = "ContextScope"
	FunctionScope SymbolScope = "FunctionScope"
)

type Symbol struct {
	Scope SymbolScope // 作用域
	Index int         // 绑定值的index
	Name  string      // 绑定变量的名称
}

type SymbolTable struct {
	store map[string]Symbol
	count int

	External *SymbolTable // 上一级
	Context  []Symbol
}

func (s *SymbolTable) Define(name string) Symbol {
	scope := GlobalScope
	if s.External != nil {
		scope = LocalScope
	}

	symbol := Symbol{
		Scope: scope,
		Index: s.count,
		Name:  name,
	}
	s.store[name] = symbol
	s.count++
	return symbol
}

func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{
		Scope: BuiltinScope,
		Index: index,
		Name:  name,
	}
	s.store[name] = symbol
	return symbol
}

func (s *SymbolTable) defineContext(ctx Symbol) Symbol {
	s.Context = append(s.Context, ctx)

	symbol := Symbol{
		Scope: ContextScope,
		Index: len(s.Context) - 1,
		Name:  ctx.Name,
	}
	s.store[ctx.Name] = symbol
	return symbol
}

func (s *SymbolTable) defineFunction(name string) Symbol {

	symbol := Symbol{
		Scope: FunctionScope,
		Name:  name,
	}
	s.store[name] = symbol
	return symbol
}

func (s *SymbolTable) GetDefine(name string) (Symbol, bool) {
	symbol, ok := s.store[name]
	if !ok && s.External != nil {
		obj, ok := s.External.GetDefine(name)
		if !ok {
			return obj, ok
		}
		if obj.Scope == GlobalScope || obj.Scope == BuiltinScope {
			return obj, ok
		}
		ctx := s.defineContext(obj)
		return ctx, true
	}
	return symbol, ok
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		store: make(map[string]Symbol),
	}
}

func NewEnclosedSymbolTable(table *SymbolTable) *SymbolTable {
	return &SymbolTable{
		store:    make(map[string]Symbol),
		External: table,
	}
}
