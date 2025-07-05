package jsonnext

import (
	"encoding/json"
	"errors"

	"github.com/google/go-jsonnet/ast"
)

func MarshalNode(node ast.Node) ([]byte, error) {
	wrappedNode := NewNode(node)
	return json.Marshal(wrappedNode)
}

func UnmarshalNode(data []byte) (ast.Node, error) {
	var node Node
	err := json.Unmarshal(data, &node)
	if err != nil {
		return nil, err
	}
	return node.Node, nil
}

type Node struct {
	Node ast.Node
}

func NewNode(node ast.Node) Node {
	return Node{Node: node}
}

func (n Node) MarshalJSON() ([]byte, error) {
	if n.Node == nil {
		return []byte("null"), nil
	}
	var proxy any
	switch v := n.Node.(type) {
	case *ast.Apply:
		proxy = Apply(*v)
	case *ast.ApplyBrace:
		proxy = ApplyBrace(*v)
	case *ast.Array:
		proxy = Array(*v)
	case *ast.ArrayComp:
		proxy = ArrayComp(*v)
	case *ast.Assert:
		proxy = Assert(*v)
	case *ast.Binary:
		proxy = Binary(*v)
	case *ast.Conditional:
		proxy = Conditional(*v)
	case *ast.Dollar:
		proxy = Dollar(*v)
	case *ast.Error:
		proxy = Error(*v)
	case *ast.Function:
		proxy = Function(*v)
	case *ast.Import:
		proxy = Import(*v)
	case *ast.ImportBin:
		proxy = ImportBin(*v)
	case *ast.ImportStr:
		proxy = ImportStr(*v)
	case *ast.InSuper:
		proxy = InSuper(*v)
	case *ast.Index:
		proxy = Index(*v)
	case *ast.LiteralBoolean:
		proxy = LiteralBoolean(*v)
	case *ast.LiteralNull:
		proxy = LiteralNull(*v)
	case *ast.LiteralNumber:
		proxy = LiteralNumber(*v)
	case *ast.LiteralString:
		proxy = LiteralString(*v)
	case *ast.Local:
		proxy = Local(*v)
	case *ast.Object:
		proxy = Object(*v)
	case *ast.ObjectComp:
		proxy = ObjectComp(*v)
	case *ast.Parens:
		proxy = Parens(*v)
	case *ast.Self:
		proxy = Self(*v)
	case *ast.Slice:
		proxy = Slice(*v)
	case *ast.SuperIndex:
		proxy = SuperIndex(*v)
	case *ast.Unary:
		proxy = Unary(*v)
	case *ast.Var:
		proxy = Var(*v)
	default:
	}
	b, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (n *Node) UnmarshalJSON(data []byte) error {
	k := struct {
		Kind string `json:"__kind__"`
	}{}
	err := json.Unmarshal(data, &k)
	if err != nil {
		return err
	}
	if string(data) == "null" {
		return nil
	}
	if k.Kind == "" {
		return errors.New("unknown Node kind")
	}
	var node ast.Node
	switch k.Kind {
	case "Apply":
		node = &Apply{}
	case "ApplyBrace":
		node = &ApplyBrace{}
	case "Array":
		node = &Array{}
	case "ArrayComp":
		node = &ArrayComp{}
	case "Assert":
		node = &Assert{}
	case "Binary":
		node = &Binary{}
	case "Conditional":
		node = &Conditional{}
	case "Dollar":
		node = &Dollar{}
	case "Error":
		node = &Error{}
	case "Function":
		node = &Function{}
	case "Import":
		node = &Import{}
	case "ImportBin":
		node = &ImportBin{}
	case "ImportStr":
		node = &ImportStr{}
	case "InSuper":
		node = &InSuper{}
	case "Index":
		node = &Index{}
	case "LiteralBoolean":
		node = &LiteralBoolean{}
	case "LiteralNull":
		node = &LiteralNull{}
	case "LiteralNumber":
		node = &LiteralNumber{}
	case "LiteralString":
		node = &LiteralString{}
	case "Local":
		node = &Local{}
	case "NodeBase":
		node = &ast.NodeBase{}
	case "Object":
		node = &Object{}
	case "ObjectComp":
		node = &ObjectComp{}
	case "Parens":
		node = &Parens{}
	case "Self":
		node = &Self{}
	case "Slice":
		node = &Slice{}
	case "SuperIndex":
		node = &SuperIndex{}
	case "Unary":
		node = &Unary{}
	case "Var":
		node = &Var{}
	default:
		// Handle unknown kind
	}
	err = json.Unmarshal(data, node)
	if err != nil {
		return err
	}

	var astNode ast.Node
	switch v := node.(type) {
	case *Apply:
		n := ast.Apply(*v)
		astNode = &n
	case *ApplyBrace:
		n := ast.ApplyBrace(*v)
		astNode = &n
	case *Array:
		n := ast.Array(*v)
		astNode = &n
	case *ArrayComp:
		n := ast.ArrayComp(*v)
		astNode = &n
	case *Assert:
		n := ast.Assert(*v)
		astNode = &n
	case *Binary:
		n := ast.Binary(*v)
		astNode = &n
	case *Conditional:
		n := ast.Conditional(*v)
		astNode = &n
	case *Dollar:
		n := ast.Dollar(*v)
		astNode = &n
	case *Error:
		n := ast.Error(*v)
		astNode = &n
	case *Function:
		n := ast.Function(*v)
		astNode = &n
	case *Import:
		n := ast.Import(*v)
		astNode = &n
	case *ImportBin:
		n := ast.ImportBin(*v)
		astNode = &n
	case *ImportStr:
		n := ast.ImportStr(*v)
		astNode = &n
	case *InSuper:
		n := ast.InSuper(*v)
		astNode = &n
	case *Index:
		n := ast.Index(*v)
		astNode = &n
	case *LiteralBoolean:
		n := ast.LiteralBoolean(*v)
		astNode = &n
	case *LiteralNull:
		n := ast.LiteralNull(*v)
		astNode = &n
	case *LiteralNumber:
		n := ast.LiteralNumber(*v)
		astNode = &n
	case *LiteralString:
		n := ast.LiteralString(*v)
		astNode = &n
	case *Local:
		n := ast.Local(*v)
		astNode = &n
	case *Object:
		n := ast.Object(*v)
		astNode = &n
	case *ObjectComp:
		n := ast.ObjectComp(*v)
		astNode = &n
	case *Parens:
		n := ast.Parens(*v)
		astNode = &n
	case *Self:
		n := ast.Self(*v)
		astNode = &n
	case *Slice:
		n := ast.Slice(*v)
		astNode = &n
	case *SuperIndex:
		n := ast.SuperIndex(*v)
		astNode = &n
	case *Unary:
		n := ast.Unary(*v)
		astNode = &n
	case *Var:
		n := ast.Var(*v)
		astNode = &n
	default:
	}
	n.Node = astNode
	return nil
}

type IfSpec ast.IfSpec

type ProxyIfSpec struct {
	Kind     string     `json:"__kind__"`
	Expr     Node       `json:"expr"`
	IfFodder ast.Fodder `json:"ifFodder"`
}

func (i IfSpec) MarshalJSON() ([]byte, error) {
	proxy := ProxyIfSpec{}
	proxy.Kind = "IfSpec"
	proxy.Expr = NewNode(i.Expr)
	proxy.IfFodder = i.IfFodder
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (i *IfSpec) UnmarshalJSON(data []byte) error {
	var proxy ProxyIfSpec
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	i.Expr = proxy.Expr.Node
	i.IfFodder = proxy.IfFodder
	return nil
}

type ForSpec ast.ForSpec

type ProxyForSpec struct {
	Kind       string         `json:"__kind__"`
	ForFodder  ast.Fodder     `json:"forFodder"`
	VarFodder  ast.Fodder     `json:"varFodder"`
	Conditions []IfSpec       `json:"conditions"`
	Outer      *ForSpec       `json:"outer"`
	Expr       Node           `json:"expr"`
	VarName    ast.Identifier `json:"varName"`
	InFodder   ast.Fodder     `json:"inFodder"`
}

func (f ForSpec) MarshalJSON() ([]byte, error) {
	proxy := ProxyForSpec{}
	proxy.Kind = "ForSpec"
	proxy.ForFodder = f.ForFodder
	proxy.VarFodder = f.VarFodder
	proxy.Conditions = make([]IfSpec, len(f.Conditions))
	for i, condition := range f.Conditions {
		proxy.Conditions[i] = IfSpec(condition)
	}
	if f.Outer != nil {
		outer := ForSpec(*f.Outer)
		proxy.Outer = &outer
	}
	proxy.Expr = NewNode(f.Expr)
	proxy.VarName = f.VarName
	proxy.InFodder = f.InFodder
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (f *ForSpec) UnmarshalJSON(data []byte) error {
	var proxy ProxyForSpec
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	f.ForFodder = proxy.ForFodder
	f.VarFodder = proxy.VarFodder
	f.Conditions = make([]ast.IfSpec, len(proxy.Conditions))
	for i, condition := range proxy.Conditions {
		f.Conditions[i] = ast.IfSpec(condition)
	}
	if proxy.Outer != nil {
		outer := ast.ForSpec(*proxy.Outer)
		f.Outer = &outer
	}
	f.Expr = proxy.Expr.Node
	f.VarName = proxy.VarName
	f.InFodder = proxy.InFodder
	return nil
}

type Apply ast.Apply

type ProxyApply struct {
	Kind             string     `json:"__kind__"`
	Target           Node       `json:"target"`
	FodderLeft       ast.Fodder `json:"fodderLeft"`
	Arguments        Arguments  `json:"arguments"`
	FodderRight      ast.Fodder `json:"fodderRight"`
	TailStrictFodder ast.Fodder `json:"tailStrictFodder"`
	ast.NodeBase
	TrailingComma bool `json:"trailingComma"`
	TailStrict    bool `json:"tailStrict"`
}

func (a Apply) MarshalJSON() ([]byte, error) {
	proxy := ProxyApply{}
	proxy.Kind = "Apply"
	proxy.Target = NewNode(a.Target)
	proxy.FodderLeft = a.FodderLeft
	proxy.Arguments = Arguments(a.Arguments)
	proxy.FodderRight = a.FodderRight
	proxy.TailStrictFodder = a.TailStrictFodder
	proxy.NodeBase = a.NodeBase
	proxy.TrailingComma = a.TrailingComma
	proxy.TailStrict = a.TailStrict
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (a *Apply) UnmarshalJSON(data []byte) error {
	var proxy ProxyApply
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	a.Target = proxy.Target.Node
	a.FodderLeft = proxy.FodderLeft
	a.Arguments = ast.Arguments(proxy.Arguments)
	a.FodderRight = proxy.FodderRight
	a.TailStrictFodder = proxy.TailStrictFodder
	a.NodeBase = proxy.NodeBase
	a.TrailingComma = proxy.TrailingComma
	a.TailStrict = proxy.TailStrict
	return nil
}

type NamedArgument ast.NamedArgument

type ProxyNamedArgument struct {
	Kind        string         `json:"__kind__"`
	NameFodder  ast.Fodder     `json:"nameFodder"`
	Name        ast.Identifier `json:"name"`
	EqFodder    ast.Fodder     `json:"eqFodder"`
	Arg         Node           `json:"arg"`
	CommaFodder ast.Fodder     `json:"commaFodder"`
}

func (n NamedArgument) MarshalJSON() ([]byte, error) {
	proxy := ProxyNamedArgument{}
	proxy.Kind = "NamedArgument"
	proxy.NameFodder = n.NameFodder
	proxy.Name = n.Name
	proxy.EqFodder = n.EqFodder
	proxy.Arg = NewNode(n.Arg)
	proxy.CommaFodder = n.CommaFodder
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (n *NamedArgument) UnmarshalJSON(data []byte) error {
	var proxy ProxyNamedArgument
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	n.NameFodder = proxy.NameFodder
	n.Name = proxy.Name
	n.EqFodder = proxy.EqFodder
	n.Arg = proxy.Arg.Node
	n.CommaFodder = proxy.CommaFodder
	return nil
}

type CommaSeparatedExpr ast.CommaSeparatedExpr

type ProxyCommaSeparatedExpr struct {
	Kind        string     `json:"__kind__"`
	Expr        Node       `json:"expr"`
	CommaFodder ast.Fodder `json:"commaFodder"`
}

func (c CommaSeparatedExpr) MarshalJSON() ([]byte, error) {
	proxy := ProxyCommaSeparatedExpr{}
	proxy.Kind = "CommaSeparatedExpr"
	proxy.Expr = NewNode(c.Expr)
	proxy.CommaFodder = c.CommaFodder
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (c *CommaSeparatedExpr) UnmarshalJSON(data []byte) error {
	var proxy ProxyCommaSeparatedExpr
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	c.Expr = proxy.Expr.Node
	c.CommaFodder = proxy.CommaFodder
	return nil
}

type Arguments ast.Arguments

type ProxyArguments struct {
	Positional []CommaSeparatedExpr `json:"positional"`
	Named      []NamedArgument      `json:"named"`
}

func (a Arguments) MarshalJSON() ([]byte, error) {
	proxy := ProxyArguments{}
	proxy.Positional = make([]CommaSeparatedExpr, len(a.Positional))
	for i, positional := range a.Positional {
		proxy.Positional[i] = CommaSeparatedExpr(positional)
	}
	proxy.Named = make([]NamedArgument, len(a.Named))
	for i, named := range a.Named {
		proxy.Named[i] = NamedArgument(named)
	}
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (a *Arguments) UnmarshalJSON(data []byte) error {
	var proxy ProxyArguments
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	a.Positional = make([]ast.CommaSeparatedExpr, len(proxy.Positional))
	for i, positional := range proxy.Positional {
		a.Positional[i] = ast.CommaSeparatedExpr(positional)
	}
	a.Named = make([]ast.NamedArgument, len(proxy.Named))
	for i, named := range proxy.Named {
		a.Named[i] = ast.NamedArgument(named)
	}
	return nil
}

type ApplyBrace ast.ApplyBrace

type ProxyApplyBrace struct {
	Kind  string `json:"__kind__"`
	Left  Node   `json:"left"`
	Right Node   `json:"right"`
	ast.NodeBase
}

func (a ApplyBrace) MarshalJSON() ([]byte, error) {
	proxy := ProxyApplyBrace{}
	proxy.Kind = "ApplyBrace"
	proxy.Left = NewNode(a.Left)
	proxy.Right = NewNode(a.Right)
	proxy.NodeBase = a.NodeBase
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (a *ApplyBrace) UnmarshalJSON(data []byte) error {
	var proxy ProxyApplyBrace
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	a.Left = proxy.Left.Node
	a.Right = proxy.Right.Node
	a.NodeBase = proxy.NodeBase
	return nil
}

type Array ast.Array

type ProxyArray struct {
	Kind        string               `json:"__kind__"`
	Elements    []CommaSeparatedExpr `json:"elements"`
	CloseFodder ast.Fodder           `json:"closeFodder"`
	ast.NodeBase
	TrailingComma bool `json:"trailingComma"`
}

func (a Array) MarshalJSON() ([]byte, error) {
	proxy := ProxyArray{}
	proxy.Kind = "Array"
	proxy.CloseFodder = a.CloseFodder
	proxy.NodeBase = a.NodeBase
	proxy.TrailingComma = a.TrailingComma
	proxy.Elements = make([]CommaSeparatedExpr, len(a.Elements))
	for i, element := range a.Elements {
		proxy.Elements[i] = CommaSeparatedExpr(element)
	}
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (a *Array) UnmarshalJSON(data []byte) error {
	var proxy ProxyArray
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	a.Elements = make([]ast.CommaSeparatedExpr, len(proxy.Elements))
	for i, element := range proxy.Elements {
		a.Elements[i] = ast.CommaSeparatedExpr(element)
	}
	a.CloseFodder = proxy.CloseFodder
	a.NodeBase = proxy.NodeBase
	a.TrailingComma = proxy.TrailingComma
	return nil
}

type ArrayComp ast.ArrayComp

type ProxyArrayComp struct {
	Kind                string     `json:"__kind__"`
	Body                Node       `json:"body"`
	TrailingCommaFodder ast.Fodder `json:"trailingCommaFodder"`
	Spec                ForSpec    `json:"spec"`
	CloseFodder         ast.Fodder `json:"closeFodder"`
	ast.NodeBase
	TrailingComma bool `json:"trailingComma"`
}

func (a ArrayComp) MarshalJSON() ([]byte, error) {
	proxy := ProxyArrayComp{}
	proxy.Kind = "ArrayComp"
	proxy.Body = NewNode(a.Body)
	proxy.TrailingCommaFodder = a.TrailingCommaFodder
	proxy.Spec = ForSpec(a.Spec)
	proxy.CloseFodder = a.CloseFodder
	proxy.NodeBase = a.NodeBase
	proxy.TrailingComma = a.TrailingComma
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (a *ArrayComp) UnmarshalJSON(data []byte) error {
	var proxy ProxyArrayComp
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	a.Body = proxy.Body.Node
	a.TrailingCommaFodder = proxy.TrailingCommaFodder
	a.Spec = ast.ForSpec(proxy.Spec)
	a.CloseFodder = proxy.CloseFodder
	a.NodeBase = proxy.NodeBase
	a.TrailingComma = proxy.TrailingComma
	return nil
}

type Assert ast.Assert

type ProxyAssert struct {
	Kind            string     `json:"__kind__"`
	Cond            Node       `json:"cond"`
	Message         Node       `json:"message"`
	Rest            Node       `json:"rest"`
	ColonFodder     ast.Fodder `json:"colonFodder"`
	SemicolonFodder ast.Fodder `json:"semicolonFodder"`
	ast.NodeBase
}

func (a Assert) MarshalJSON() ([]byte, error) {
	proxy := ProxyAssert{}
	proxy.Kind = "Assert"
	proxy.Cond = NewNode(a.Cond)
	proxy.Message = NewNode(a.Message)
	proxy.Rest = NewNode(a.Rest)
	proxy.ColonFodder = a.ColonFodder
	proxy.SemicolonFodder = a.SemicolonFodder
	proxy.NodeBase = a.NodeBase
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (a *Assert) UnmarshalJSON(data []byte) error {
	var proxy ProxyAssert
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	a.Cond = proxy.Cond.Node
	a.Message = proxy.Message.Node
	a.Rest = proxy.Rest.Node
	a.ColonFodder = proxy.ColonFodder
	a.SemicolonFodder = proxy.SemicolonFodder
	a.NodeBase = proxy.NodeBase
	return nil
}

type Binary ast.Binary

type ProxyBinary struct {
	Kind     string     `json:"__kind__"`
	Right    Node       `json:"right"`
	Left     Node       `json:"left"`
	OpFodder ast.Fodder `json:"opFodder"`
	ast.NodeBase
	Op ast.BinaryOp `json:"op"`
}

func (b Binary) MarshalJSON() ([]byte, error) {
	proxy := ProxyBinary{}
	proxy.Kind = "Binary"
	proxy.Right = NewNode(b.Right)
	proxy.Left = NewNode(b.Left)
	proxy.OpFodder = b.OpFodder
	proxy.NodeBase = b.NodeBase
	proxy.Op = b.Op
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (b *Binary) UnmarshalJSON(data []byte) error {
	var proxy ProxyBinary
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	b.Right = proxy.Right.Node
	b.Left = proxy.Left.Node
	b.OpFodder = proxy.OpFodder
	b.NodeBase = proxy.NodeBase
	b.Op = proxy.Op
	return nil
}

type Conditional ast.Conditional

type ProxyConditional struct {
	Kind        string     `json:"__kind__"`
	Cond        Node       `json:"cond"`
	BranchTrue  Node       `json:"branchTrue"`
	BranchFalse Node       `json:"branchFalse"`
	ThenFodder  ast.Fodder `json:"thenFodder"`
	ElseFodder  ast.Fodder `json:"elseFodder"`
	ast.NodeBase
}

func (c Conditional) MarshalJSON() ([]byte, error) {
	proxy := ProxyConditional{}
	proxy.Kind = "Conditional"
	proxy.Cond = NewNode(c.Cond)
	proxy.BranchTrue = NewNode(c.BranchTrue)
	proxy.BranchFalse = NewNode(c.BranchFalse)
	proxy.ThenFodder = c.ThenFodder
	proxy.ElseFodder = c.ElseFodder
	proxy.NodeBase = c.NodeBase
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (c *Conditional) UnmarshalJSON(data []byte) error {
	var proxy ProxyConditional
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	c.Cond = proxy.Cond.Node
	c.BranchTrue = proxy.BranchTrue.Node
	c.BranchFalse = proxy.BranchFalse.Node
	c.ThenFodder = proxy.ThenFodder
	c.ElseFodder = proxy.ElseFodder
	c.NodeBase = proxy.NodeBase
	return nil
}

type Dollar ast.Dollar

type ProxyDollar struct {
	Kind string `json:"__kind__"`
	ast.NodeBase
}

func (d Dollar) MarshalJSON() ([]byte, error) {
	proxy := ProxyDollar{}
	proxy.Kind = "Dollar"
	proxy.NodeBase = d.NodeBase
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (d *Dollar) UnmarshalJSON(data []byte) error {
	var proxy ProxyDollar
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	d.NodeBase = proxy.NodeBase
	return nil
}

type Error ast.Error

type ProxyError struct {
	Kind string `json:"__kind__"`
	Expr Node   `json:"expr"`
	ast.NodeBase
}

func (e Error) MarshalJSON() ([]byte, error) {
	proxy := ProxyError{}
	proxy.Kind = "Error"
	proxy.Expr = NewNode(e.Expr)
	proxy.NodeBase = e.NodeBase
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (e *Error) UnmarshalJSON(data []byte) error {
	var proxy ProxyError
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	e.Expr = proxy.Expr.Node
	e.NodeBase = proxy.NodeBase
	return nil
}

type Function ast.Function

type ProxyFunction struct {
	Kind             string      `json:"__kind__"`
	ParenLeftFodder  ast.Fodder  `json:"parenLeftFodder"`
	ParenRightFodder ast.Fodder  `json:"parenRightFodder"`
	Body             Node        `json:"body"`
	Parameters       []Parameter `json:"parameters"`
	ast.NodeBase
	TrailingComma bool `json:"trailingComma"`
}

func (f Function) MarshalJSON() ([]byte, error) {
	proxy := ProxyFunction{}
	proxy.Kind = "Function"
	proxy.ParenLeftFodder = f.ParenLeftFodder
	proxy.ParenRightFodder = f.ParenRightFodder
	proxy.Body = NewNode(f.Body)
	proxy.NodeBase = f.NodeBase
	proxy.TrailingComma = f.TrailingComma
	proxy.Parameters = make([]Parameter, len(f.Parameters))
	for i, parameter := range f.Parameters {
		proxy.Parameters[i] = Parameter(parameter)
	}
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (f *Function) UnmarshalJSON(data []byte) error {
	var proxy ProxyFunction
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	f.ParenLeftFodder = proxy.ParenLeftFodder
	f.ParenRightFodder = proxy.ParenRightFodder
	f.Body = proxy.Body.Node
	f.Parameters = make([]ast.Parameter, len(proxy.Parameters))
	for i, parameter := range proxy.Parameters {
		f.Parameters[i] = ast.Parameter(parameter)
	}
	f.NodeBase = proxy.NodeBase
	f.TrailingComma = proxy.TrailingComma
	return nil
}

type Parameter ast.Parameter

type ProxyParameter struct {
	Kind        string            `json:"__kind__"`
	NameFodder  ast.Fodder        `json:"nameFodder"`
	Name        ast.Identifier    `json:"name"`
	CommaFodder ast.Fodder        `json:"commaFodder"`
	EqFodder    ast.Fodder        `json:"eqFodder"`
	DefaultArg  Node              `json:"defaultArg"`
	LocRange    ast.LocationRange `json:"locRange"`
}

func (p Parameter) MarshalJSON() ([]byte, error) {
	proxy := ProxyParameter{}
	proxy.Kind = "Parameter"
	proxy.NameFodder = p.NameFodder
	proxy.Name = p.Name
	proxy.CommaFodder = p.CommaFodder
	proxy.EqFodder = p.EqFodder
	proxy.DefaultArg = NewNode(p.DefaultArg)
	proxy.LocRange = p.LocRange
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (p *Parameter) UnmarshalJSON(data []byte) error {
	var proxy ProxyParameter
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	p.NameFodder = proxy.NameFodder
	p.Name = proxy.Name
	p.CommaFodder = proxy.CommaFodder
	p.EqFodder = proxy.EqFodder
	p.DefaultArg = proxy.DefaultArg.Node
	p.LocRange = proxy.LocRange
	return nil
}

type Import ast.Import

type ProxyImport struct {
	Kind string             `json:"__kind__"`
	File *ast.LiteralString `json:"file"`
	ast.NodeBase
}

func (i Import) MarshalJSON() ([]byte, error) {
	proxy := ProxyImport{}
	proxy.Kind = "Import"
	proxy.File = i.File
	proxy.NodeBase = i.NodeBase
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (i *Import) UnmarshalJSON(data []byte) error {
	var proxy ProxyImport
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	i.File = proxy.File
	i.NodeBase = proxy.NodeBase
	return nil
}

type ImportBin ast.ImportBin

type ProxyImportBin struct {
	Kind string             `json:"__kind__"`
	File *ast.LiteralString `json:"file"`
	ast.NodeBase
}

func (i ImportBin) MarshalJSON() ([]byte, error) {
	proxy := ProxyImportBin{}
	proxy.Kind = "ImportBin"
	proxy.File = i.File
	proxy.NodeBase = i.NodeBase
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (i *ImportBin) UnmarshalJSON(data []byte) error {
	var proxy ProxyImportBin
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	i.File = proxy.File
	i.NodeBase = proxy.NodeBase
	return nil
}

type ImportStr ast.ImportStr

type ProxyImportStr struct {
	Kind string             `json:"__kind__"`
	File *ast.LiteralString `json:"file"`
	ast.NodeBase
}

func (i ImportStr) MarshalJSON() ([]byte, error) {
	proxy := ProxyImportStr{}
	proxy.Kind = "ImportStr"
	proxy.File = i.File
	proxy.NodeBase = i.NodeBase
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (i *ImportStr) UnmarshalJSON(data []byte) error {
	var proxy ProxyImportStr
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	i.File = proxy.File
	i.NodeBase = proxy.NodeBase
	return nil
}

type Index ast.Index

type ProxyIndex struct {
	Kind               string          `json:"__kind__"`
	Target             Node            `json:"target"`
	Index              Node            `json:"index"`
	RightBracketFodder ast.Fodder      `json:"rightBracketFodder"`
	LeftBracketFodder  ast.Fodder      `json:"leftBracketFodder"`
	Id                 *ast.Identifier `json:"id"`
	ast.NodeBase
}

func (i Index) MarshalJSON() ([]byte, error) {
	proxy := ProxyIndex{}
	proxy.Kind = "Index"
	proxy.Target = NewNode(i.Target)
	proxy.Index = NewNode(i.Index)
	proxy.RightBracketFodder = i.RightBracketFodder
	proxy.LeftBracketFodder = i.LeftBracketFodder
	proxy.Id = i.Id
	proxy.NodeBase = i.NodeBase
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (i *Index) UnmarshalJSON(data []byte) error {
	var proxy ProxyIndex
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	i.Target = proxy.Target.Node
	i.Index = proxy.Index.Node
	i.RightBracketFodder = proxy.RightBracketFodder
	i.LeftBracketFodder = proxy.LeftBracketFodder
	i.Id = proxy.Id
	i.NodeBase = proxy.NodeBase
	return nil
}

type LiteralBoolean ast.LiteralBoolean

type ProxyLiteralBoolean struct {
	Kind  string `json:"__kind__"`
	Value bool   `json:"value"`
	ast.NodeBase
}

func (l LiteralBoolean) MarshalJSON() ([]byte, error) {
	proxy := ProxyLiteralBoolean{}
	proxy.Kind = "LiteralBoolean"
	proxy.Value = l.Value
	proxy.NodeBase = l.NodeBase
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (l *LiteralBoolean) UnmarshalJSON(data []byte) error {
	var proxy ProxyLiteralBoolean
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	l.Value = proxy.Value
	l.NodeBase = proxy.NodeBase
	return nil
}

type LiteralNull ast.LiteralNull

type ProxyLiteralNull struct {
	Kind string `json:"__kind__"`
	ast.NodeBase
}

func (l LiteralNull) MarshalJSON() ([]byte, error) {
	proxy := ProxyLiteralNull{}
	proxy.Kind = "LiteralNull"
	proxy.NodeBase = l.NodeBase
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (l *LiteralNull) UnmarshalJSON(data []byte) error {
	var proxy ProxyLiteralNull
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	l.NodeBase = proxy.NodeBase
	return nil
}

type LiteralString ast.LiteralString

type ProxyLiteralString struct {
	NodeKind        string `json:"__kind__"`
	Value           string `json:"value"`
	BlockIndent     string `json:"blockIndent"`
	BlockTermIndent string `json:"blockTermIndent"`
	ast.NodeBase
	Kind ast.LiteralStringKind `json:"kind"`
}

func (l LiteralString) MarshalJSON() ([]byte, error) {
	proxy := ProxyLiteralString{}
	proxy.NodeKind = "LiteralString"
	proxy.Value = l.Value
	proxy.NodeBase = l.NodeBase
	proxy.Kind = l.Kind
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (l *LiteralString) UnmarshalJSON(data []byte) error {
	var proxy ProxyLiteralString
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	l.Value = proxy.Value
	l.NodeBase = proxy.NodeBase
	return nil
}

type LiteralNumber ast.LiteralNumber

type ProxyLiteralNumber struct {
	Kind           string `json:"__kind__"`
	OriginalString string `json:"originalString"`
	ast.NodeBase
}

func (l LiteralNumber) MarshalJSON() ([]byte, error) {
	proxy := ProxyLiteralNumber{}
	proxy.Kind = "LiteralNumber"
	proxy.OriginalString = l.OriginalString
	proxy.NodeBase = l.NodeBase
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (l *LiteralNumber) UnmarshalJSON(data []byte) error {
	var proxy ProxyLiteralNumber
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	l.OriginalString = proxy.OriginalString
	l.NodeBase = proxy.NodeBase
	return nil
}

type Slice ast.Slice

type ProxySlice struct {
	Kind               string     `json:"__kind__"`
	Target             Node       `json:"target"`
	LeftBracketFodder  ast.Fodder `json:"leftBracketFodder"`
	BeginIndex         Node       `json:"beginIndex"`
	EndColonFodder     ast.Fodder `json:"endColonFodder"`
	EndIndex           Node       `json:"endIndex"`
	StepColonFodder    ast.Fodder `json:"stepColonFodder"`
	Step               Node       `json:"step"`
	RightBracketFodder ast.Fodder `json:"rightBracketFodder"`
	ast.NodeBase
}

func (s Slice) MarshalJSON() ([]byte, error) {
	proxy := ProxySlice{}
	proxy.Kind = "Slice"
	proxy.Target = NewNode(s.Target)
	proxy.LeftBracketFodder = s.LeftBracketFodder
	proxy.BeginIndex = NewNode(s.BeginIndex)
	proxy.EndColonFodder = s.EndColonFodder
	proxy.EndIndex = NewNode(s.EndIndex)
	proxy.StepColonFodder = s.StepColonFodder
	proxy.Step = NewNode(s.Step)
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (s *Slice) UnmarshalJSON(data []byte) error {
	var proxy ProxySlice
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	s.Target = proxy.Target.Node
	s.LeftBracketFodder = proxy.LeftBracketFodder
	s.BeginIndex = proxy.BeginIndex.Node
	s.EndColonFodder = proxy.EndColonFodder
	s.EndIndex = proxy.EndIndex.Node
	s.StepColonFodder = proxy.StepColonFodder
	s.Step = proxy.Step.Node
	s.RightBracketFodder = proxy.RightBracketFodder
	s.NodeBase = proxy.NodeBase
	return nil
}

type LocalBind ast.LocalBind

type ProxyLocalBind struct {
	Kind        string         `json:"__kind__"`
	VarFodder   ast.Fodder     `json:"varFodder"`
	Body        Node           `json:"body"`
	EqFodder    ast.Fodder     `json:"eqFodder"`
	Variable    ast.Identifier `json:"variable"`
	CloseFodder ast.Fodder     `json:"closeFodder"`
	Fun         *Function      `json:"fun"`
}

func (l LocalBind) MarshalJSON() ([]byte, error) {
	proxy := ProxyLocalBind{}
	proxy.Kind = "LocalBind"
	proxy.VarFodder = l.VarFodder
	proxy.Body = NewNode(l.Body)
	proxy.EqFodder = l.EqFodder
	proxy.Variable = l.Variable
	proxy.CloseFodder = l.CloseFodder
	if l.Fun != nil {
		fun := Function(*l.Fun)
		proxy.Fun = &fun
	}
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (l *LocalBind) UnmarshalJSON(data []byte) error {
	var proxy ProxyLocalBind
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	l.VarFodder = proxy.VarFodder
	l.Body = proxy.Body.Node
	l.EqFodder = proxy.EqFodder
	l.Variable = proxy.Variable
	l.CloseFodder = proxy.CloseFodder
	if proxy.Fun != nil {
		fun := ast.Function(*proxy.Fun)
		l.Fun = &fun
	}
	return nil
}

type ProxyLocal struct {
	Kind  string      `json:"__kind__"`
	Binds []LocalBind `json:"binds"`
	Body  Node        `json:"body"`
	ast.NodeBase
}

type Local ast.Local

func (l Local) MarshalJSON() ([]byte, error) {
	proxy := ProxyLocal{}
	proxy.Kind = "Local"
	proxy.Binds = make([]LocalBind, len(l.Binds))
	for i, bind := range l.Binds {
		proxy.Binds[i] = LocalBind(bind)
	}
	proxy.Body = NewNode(l.Body)
	proxy.NodeBase = l.NodeBase
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (l *Local) UnmarshalJSON(data []byte) error {
	var proxy ProxyLocal
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	l.Binds = make([]ast.LocalBind, len(proxy.Binds))
	for i, bind := range proxy.Binds {
		l.Binds[i] = ast.LocalBind(bind)
	}
	l.Body = proxy.Body.Node
	l.NodeBase = proxy.NodeBase
	return nil
}

type ObjectField ast.ObjectField

type ProxyObjectField struct {
	NodeKind    string              `json:"__kind__"`
	Method      *Function           `json:"method"`
	Id          *ast.Identifier     `json:"id"`
	Fodder2     ast.Fodder          `json:"fodder2"`
	Fodder1     ast.Fodder          `json:"fodder1"`
	OpFodder    ast.Fodder          `json:"opFodder"`
	CommaFodder ast.Fodder          `json:"commaFodder"`
	Expr1       Node                `json:"expr1"`
	Expr2       Node                `json:"expr2"`
	Expr3       Node                `json:"expr3"`
	LocRange    ast.LocationRange   `json:"locRange"`
	Kind        ast.ObjectFieldKind `json:"kind"`
	Hide        ast.ObjectFieldHide
	SuperSugar  bool
}

func (o ObjectField) MarshalJSON() ([]byte, error) {
	proxy := ProxyObjectField{}
	proxy.NodeKind = "ObjectField"
	if o.Method != nil {
		method := Function(*o.Method)
		proxy.Method = &method
	}
	proxy.Id = o.Id
	proxy.Fodder2 = o.Fodder2
	proxy.Fodder1 = o.Fodder1
	proxy.OpFodder = o.OpFodder
	proxy.CommaFodder = o.CommaFodder
	proxy.Expr1 = NewNode(o.Expr1)
	proxy.Expr2 = NewNode(o.Expr2)
	proxy.Expr3 = NewNode(o.Expr3)
	proxy.LocRange = o.LocRange
	proxy.Kind = o.Kind
	proxy.Hide = o.Hide
	proxy.SuperSugar = o.SuperSugar
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (o *ObjectField) UnmarshalJSON(data []byte) error {
	var proxy ProxyObjectField
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	if proxy.Method != nil {
		method := ast.Function(*proxy.Method)
		o.Method = &method
	}
	o.Id = proxy.Id
	o.Fodder2 = proxy.Fodder2
	o.Fodder1 = proxy.Fodder1
	o.OpFodder = proxy.OpFodder
	o.CommaFodder = proxy.CommaFodder
	o.Expr1 = proxy.Expr1.Node
	o.Expr2 = proxy.Expr2.Node
	o.Expr3 = proxy.Expr3.Node
	o.LocRange = proxy.LocRange
	o.Kind = proxy.Kind
	o.Hide = proxy.Hide
	o.SuperSugar = proxy.SuperSugar
	return nil
}

type Object ast.Object

type ProxyObject struct {
	Kind        string        `json:"__kind__"`
	Fields      []ObjectField `json:"fields"`
	CloseFodder ast.Fodder    `json:"closeFodder"`
	ast.NodeBase
	TrailingComma bool `json:"trailingComma"`
}

func (o Object) MarshalJSON() ([]byte, error) {
	proxy := ProxyObject{}
	proxy.Kind = "Object"
	proxy.Fields = make([]ObjectField, len(o.Fields))
	for i, field := range o.Fields {
		proxy.Fields[i] = ObjectField(field)
	}
	proxy.CloseFodder = o.CloseFodder
	proxy.NodeBase = o.NodeBase
	proxy.TrailingComma = o.TrailingComma
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (o *Object) UnmarshalJSON(data []byte) error {
	var proxy ProxyObject
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	o.Fields = make([]ast.ObjectField, len(proxy.Fields))
	for i, field := range proxy.Fields {
		o.Fields[i] = ast.ObjectField(field)
	}
	o.CloseFodder = proxy.CloseFodder
	o.NodeBase = proxy.NodeBase
	o.TrailingComma = proxy.TrailingComma
	return nil
}

type ObjectComp ast.ObjectComp

type ProxyObjectComp struct {
	Kind                string        `json:"__kind__"`
	Fields              []ObjectField `json:"fields"`
	TrailingCommaFodder ast.Fodder    `json:"trailingCommaFodder"`
	CloseFodder         ast.Fodder    `json:"closeFodder"`
	Spec                ForSpec       `json:"spec"`
	ast.NodeBase
	TrailingComma bool `json:"trailingComma"`
}

func (o ObjectComp) MarshalJSON() ([]byte, error) {
	proxy := ProxyObjectComp{}
	proxy.Kind = "ObjectComp"
	proxy.Fields = make([]ObjectField, len(o.Fields))
	for i, field := range o.Fields {
		proxy.Fields[i] = ObjectField(field)
	}
	proxy.TrailingCommaFodder = o.TrailingCommaFodder
	proxy.CloseFodder = o.CloseFodder
	proxy.Spec = ForSpec(o.Spec)
	proxy.NodeBase = o.NodeBase
	proxy.TrailingComma = o.TrailingComma
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (o *ObjectComp) UnmarshalJSON(data []byte) error {
	var proxy ProxyObjectComp
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	o.Fields = make([]ast.ObjectField, len(proxy.Fields))
	for i, field := range proxy.Fields {
		o.Fields[i] = ast.ObjectField(field)
	}
	o.TrailingCommaFodder = proxy.TrailingCommaFodder
	o.CloseFodder = proxy.CloseFodder
	o.Spec = ast.ForSpec(proxy.Spec)
	o.NodeBase = proxy.NodeBase
	o.TrailingComma = proxy.TrailingComma
	return nil
}

type Parens ast.Parens

type ProxyParens struct {
	Kind        string     `json:"__kind__"`
	Inner       Node       `json:"inner"`
	CloseFodder ast.Fodder `json:"closeFodder"`
	ast.NodeBase
}

func (p Parens) MarshalJSON() ([]byte, error) {
	proxy := ProxyParens{}
	proxy.Kind = "Parens"
	proxy.Inner = NewNode(p.Inner)
	proxy.CloseFodder = p.CloseFodder
	proxy.NodeBase = p.NodeBase
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (p *Parens) UnmarshalJSON(data []byte) error {
	var proxy ProxyParens
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	p.Inner = proxy.Inner.Node
	p.CloseFodder = proxy.CloseFodder
	p.NodeBase = proxy.NodeBase
	return nil
}

type Self ast.Self

type ProxySelf struct {
	Kind string `json:"__kind__"`
	ast.NodeBase
}

func (s Self) MarshalJSON() ([]byte, error) {
	proxy := ProxySelf{}
	proxy.Kind = "Self"
	proxy.NodeBase = s.NodeBase
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (s *Self) UnmarshalJSON(data []byte) error {
	var proxy ProxySelf
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	s.NodeBase = proxy.NodeBase
	return nil
}

type SuperIndex ast.SuperIndex

type ProxySuperIndex struct {
	Kind      string          `json:"__kind__"`
	IDFodder  ast.Fodder      `json:"idFodder"`
	Index     Node            `json:"index"`
	DotFodder ast.Fodder      `json:"dotFodder"`
	Id        *ast.Identifier `json:"id"`
	ast.NodeBase
}

func (s SuperIndex) MarshalJSON() ([]byte, error) {
	proxy := ProxySuperIndex{}
	proxy.Kind = "SuperIndex"
	proxy.IDFodder = s.IDFodder
	proxy.Index = NewNode(s.Index)
	proxy.DotFodder = s.DotFodder
	proxy.Id = s.Id
	proxy.NodeBase = s.NodeBase
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (s *SuperIndex) UnmarshalJSON(data []byte) error {
	var proxy ProxySuperIndex
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	s.IDFodder = proxy.IDFodder
	s.Index = proxy.Index.Node
	s.DotFodder = proxy.DotFodder
	s.Id = proxy.Id
	s.NodeBase = proxy.NodeBase
	return nil
}

type InSuper ast.InSuper

type ProxyInSuper struct {
	Kind        string     `json:"__kind__"`
	Index       Node       `json:"index"`
	InFodder    ast.Fodder `json:"inFodder"`
	SuperFodder ast.Fodder `json:"superFodder"`
	ast.NodeBase
}

func (i InSuper) MarshalJSON() ([]byte, error) {
	proxy := ProxyInSuper{}
	proxy.Kind = "InSuper"
	proxy.Index = NewNode(i.Index)
	proxy.InFodder = i.InFodder
	proxy.SuperFodder = i.SuperFodder
	proxy.NodeBase = i.NodeBase
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (i *InSuper) UnmarshalJSON(data []byte) error {
	var proxy ProxyInSuper
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	i.Index = proxy.Index.Node
	i.InFodder = proxy.InFodder
	i.SuperFodder = proxy.SuperFodder
	i.NodeBase = proxy.NodeBase
	return nil
}

type Unary ast.Unary

type ProxyUnary struct {
	Kind string `json:"__kind__"`
	Expr Node   `json:"expr"`
	ast.NodeBase
	Op ast.UnaryOp `json:"op"`
}

func (u Unary) MarshalJSON() ([]byte, error) {
	proxy := ProxyUnary{}
	proxy.Kind = "Unary"
	proxy.Expr = NewNode(u.Expr)
	proxy.NodeBase = u.NodeBase
	proxy.Op = u.Op
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (u *Unary) UnmarshalJSON(data []byte) error {
	var proxy ProxyUnary
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	u.Expr = proxy.Expr.Node
	u.NodeBase = proxy.NodeBase
	u.Op = proxy.Op
	return nil
}

type Var ast.Var

type ProxyVar struct {
	Kind string         `json:"__kind__"`
	Id   ast.Identifier `json:"id"`
	ast.NodeBase
}

func (v Var) MarshalJSON() ([]byte, error) {
	proxy := ProxyVar{}
	proxy.Kind = "Var"
	proxy.Id = v.Id
	proxy.NodeBase = v.NodeBase
	j, err := json.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (v *Var) UnmarshalJSON(data []byte) error {
	var proxy ProxyVar
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}
	v.Id = proxy.Id
	v.NodeBase = proxy.NodeBase
	return nil
}
