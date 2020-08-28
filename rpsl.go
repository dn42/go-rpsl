package rpsl

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

type RPSL struct {
	Schema map[string]*Schema

	index Indexer
	fetch Fetcher
}

func NewRPSL(opts ...Option) *RPSL {
	rpsl := &RPSL{}

	for _, o := range opts {
		o.Apply(rpsl)
	}

	if rpsl.Schema == nil {
		rpsl.Schema = make(map[string]*Schema)
	}

	if rpsl.fetch == nil {
		rpsl.fetch = &nullFS{}
	}

	if rpsl.index == nil {
		rpsl.index = &nullFS{}
	}

	return rpsl
}

type Option interface {
	Apply(*RPSL)
}

type Fetcher interface {
	LoadObject(schema, name string) (*Object, error)
}

type Indexer interface {
	FindObject(search string) ([]*Object, error)
}

var PadLength = 19

type Object struct {
	attributes ListAttribute
	keys       map[string][]int
	schema     *Schema
}

func ParseObject(content string) *Object {
	r := strings.NewReader(content)
	p := NewParser(r)
	p.Scan()

	return p.Current()
}

func (dom *Object) Name() string {
	if dom == nil || len(dom.attributes) == 0 {
		return ""
	}

	fields := dom.Get(dom.Primary()).Fields()
	if dom == nil || len(fields) == 0 {
		return ""
	}

	return fields[0]
}

func (dom *Object) Primary() string {
	if dom != nil && dom.schema != nil {
		return dom.schema.Primary
	}
	return dom.Schema()
}

func (dom *Object) String() string {
	padLen := PadLength
	for attr := range dom.keys {
		if len(attr) > padLen {
			padLen = len(attr) + 2
		}
	}
	var lis []string
	for _, attr := range dom.attributes {
		if attr == nil {
			continue
		}
		lis = append(lis, attr.StringN(padLen))
	}

	return strings.Join(lis, "\n")
}

func (dom *Object) Get(name string) *Attribute {
	return dom.GetN(name, 0)
}

func (dom *Object) GetN(name string, index int) *Attribute {
	if key, ok := dom.keys[name]; ok {
		if len(key) > index {
			return dom.Attr(key[index])
		}
	}
	return nil
}

func (dom *Object) GetAll(name string) ListAttribute {
	if keys, ok := dom.keys[name]; ok {
		lis := make([]*Attribute, len(keys))
		for i, key := range keys {
			lis[i] = dom.Attr(key)
		}
		return lis
	}

	return nil
}

func (dom *Object) Set(name string, values ...string) {
	dom.SetN(name, 0, values...)
}

func (dom *Object) SetN(name string, index int, values ...string) {
	var rows []Value
	switch len(values) {
	case 0:
	case 1:
		rows = append(rows, Value{Value: values[0]})
	default:
		rows = make([]Value, len(values))
		for i, v := range values {
			rows[i] = Value{Value: v}
		}
	}

	if key, ok := dom.keys[name]; ok {
		if index > -1 && len(key) > index {
			dom.attributes[key[index]].rows = rows
		} else {
			dom.keys[name] = append(key, len(dom.attributes))
			dom.attributes = append(dom.attributes, &Attribute{Name: name, rows: rows})
		}
	}
}

func (dom *Object) Add(name string, values ...string) {
	var rows []Value
	switch len(values) {
	case 0:
	case 1:
		rows = append(rows, NewValue(values[0]))
	default:
		rows = make([]Value, len(values))
		for i, v := range values {
			rows[i] = NewValue(v)
		}
	}

	if key, ok := dom.keys[name]; ok {
		dom.keys[name] = append(key, len(dom.attributes))
	} else {
		dom.keys[name] = []int{len(dom.attributes)}
	}
	dom.attributes = append(dom.attributes, &Attribute{Name: name, rows: rows})
}

func (dom *Object) Delete(name string) {
	if k, ok := dom.keys[name]; ok {
		if len(k) > 0 {
			dom.attributes[k[0]] = nil
			k = k[1:]
		}
		if len(k) == 0 {
			delete(dom.keys, name)
		}
	}
}

func (dom *Object) Schema() string {
	if dom == nil || len(dom.attributes) == 0 {
		return ""
	}
	return dom.attributes[0].Name
}

func (dom *Object) Attr(index int) *Attribute {
	if index > len(dom.attributes) {
		panic("index out of range")
	}
	a := dom.attributes[index]
	attr := &Attribute{Name: a.Name, rows: make([]Value, len(a.rows))}
	copy(attr.rows, a.rows)
	if dom.schema != nil {
		attr.spec = dom.schema.Spec(a.Name)
	}

	return attr
}

func (dom *Object) Attrs() ListAttribute {
	lis := make([]*Attribute, len(dom.attributes))
	for i := range dom.attributes {
		lis[i] = dom.Attr(i)
	}

	return lis
}

func (dom *Object) MarshalJSON() ([]byte, error) {
	return json.Marshal(dom.Attrs())
}

type ListObject []*Object

func (lis ListObject) String() string {
	if len(lis) == 0 {
		return ""
	}

	s := lis[0].String()

	if len(lis) == 1 {
		return s
	}

	var buf strings.Builder
	buf.WriteString(s)

	for _, dom := range lis[1:] {
		buf.WriteRune('\n')
		buf.WriteRune('\n')
		buf.WriteString(dom.String())
	}

	return buf.String()
}

func ParseAll(in io.Reader) ListObject {
	p := NewParser(in)
	var lis []*Object
	for p.Scan() {
		lis = append(lis, p.Current())
	}
	return lis
}

type Attribute struct {
	Name string
	rows []Value
	spec Spec
}

func (attr *Attribute) Lines() []string {
	if attr == nil || len(attr.rows) == 0 {
		return nil
	}

	lis := make([]string, len(attr.rows))
	for i, row := range attr.rows {
		lis[i] = row.Value
	}

	return lis
}

func (attr *Attribute) Text() string {
	lis := attr.Lines()

	switch len(lis) {
	case 0:
		return ""
	case 1:
		return lis[0]
	default:
		return strings.Join(lis, "\n")
	}
}

func (attr *Attribute) Comment() string {
	lis := make([]string, 0, len(attr.rows))
	for _, v := range attr.rows {
		if c := v.Comment; len(c) > 0 {
			lis = append(lis, c)
		}
	}
	return strings.Join(lis, "\n")
}

func (attr *Attribute) Default(text string) string {
	if attr == nil {
		return text
	}

	return attr.Text()
}

func (attr *Attribute) Fields() []string {
	return strings.Fields(attr.Text())
}

func (attr *Attribute) StringN(pad int) string {
	if attr == nil {
		return ""
	}

	if len(attr.rows) == 0 {
		return attr.Name + ":" + strings.Repeat(" ", pad-len(attr.Name))
	}

	var lis []string
	s := attr.Name + ":" + strings.Repeat(" ", pad-len(attr.Name)) + attr.rows[0].String()
	lis = append(lis, s)
	for _, row := range attr.rows[1:] {
		v := row.String()

		if v == "" {
			lis = append(lis, "+")

			continue
		}
		lis = append(lis, strings.Repeat(" ", pad+1)+v)
	}

	return strings.Join(lis, "\n")
}

func (attr *Attribute) String() string {
	return attr.StringN(19)
}

func (attr *Attribute) Raw() string {
	lis := make([]string, len(attr.rows))
	for i, row := range attr.rows {
		lis[i] = row.String()
	}

	return strings.Join(lis, "\n")
}

func (attr *Attribute) Args() *Arguments {
	fields := attr.Fields()
	args := NewArguments()

	if attr == nil {
		return args
	}

	spec := append(attr.spec, &SpecRuleText{})

	i := 0
	for _, s := range spec {
		if i >= len(fields) {
			break
		}
		i += s.ApplyArgument(args, fields[i:])
	}

	return args
}

func (attr *Attribute) MarshalJSON() ([]byte, error) {
	lis := make([]interface{}, 2)
	lis[0] = attr.Name
	if attr.spec != nil {
		lis[1] = attr.Args()
	} else {
		lis[1] = attr.Raw()
	}

	return json.Marshal(lis)
}

type ListAttribute []*Attribute

func (lis ListAttribute) Text() string {
	switch len(lis) {
	case 0:
		return ""
	case 1:
		return lis[0].Text()
	default:
		arr := make([]string, len(lis))
		for i, a := range lis {
			arr[i] = a.Text()
		}
		return strings.Join(arr, "\n")
	}
}
func (lis ListAttribute) Default(text string) string {
	if lis == nil || len(lis) == 0 {
		return text
	}

	return lis.Text()
}
func (lis ListAttribute) Fields() []string {
	return strings.Fields(lis.Text())
}

type Value struct {
	Lineno  int
	Value   string
	Comment string
}

func NewValue(s string) Value {
	sp := strings.SplitN(s, "#", 2)
	v := Value{Value: strings.TrimSpace(sp[0])}
	if len(sp) > 1 {
		v.Comment = strings.TrimSpace(sp[1])
	}
	return v
}

func (v Value) String() string {
	if len(v.Comment) > 0 {
		if len(v.Value) == 0 {
			return "# " + v.Comment
		}

		return v.Value + " # " + v.Comment
	}

	return v.Value
}

type Schema struct {
	*Object

	Name    string
	Primary string
	Links   map[string][]string
	specTx  map[string][]string
	spec    map[string]Spec
	Rules   map[string]*Set
}

func (s *Schema) Spec(name string) Spec {
	if spec, ok := s.spec[name]; ok {
		lis := make([]SpecRule, len(spec))
		copy(lis, spec)
		return lis
	}

	return nil
}
func (s *Schema) String() string {
	var buf strings.Builder

	buf.WriteString("schema: ")
	buf.WriteString(s.Name)
	buf.WriteRune('\n')
	buf.WriteString("primary: ")
	buf.WriteString(s.Primary)
	for key, rules := range s.Rules {
		buf.WriteRune('\n')
		buf.WriteString(key)
		buf.WriteRune(':')
		buf.WriteRune(' ')
		buf.WriteString(rules.String())
	}

	return buf.String()
}

type Schemas struct {
	m map[string]*Schema
}

func ParseSchemas(lis ListObject) (*Schemas, error) {
	p := &SchemaParser{keys: NewSet()}

	// Second Pass: Parse objects
	schemas := &Schemas{m: make(map[string]*Schema)}
	for _, dom := range lis {
		if dom.Schema() != "schema" {
			continue
		}

		schema := p.ParseSchema(dom)
		schemas.m[schema.Name] = schema
		p.keys.Add(schema.Primary)
	}

	var err error
	for schemaName, schema := range schemas.m {
		for name, rule := range schema.specTx {
			schema.spec[name], err = p.ParseSpec(rule)
			if err != nil {
				return nil, fmt.Errorf("parsing schema %s key %s: %w", schemaName, name, err)
			}
		}
	}

	return schemas, err
}
func (s *Schemas) Apply(lis ...*Object) {
	for _, dom := range lis {
		if schema, ok := s.m[dom.Schema()]; ok {
			dom.schema = schema
		}
	}
}
func (s *Schemas) Items() []*Schema {
	lis := make([]*Schema, 0, len(s.m))
	for _, schema := range s.m {
		lis = append(lis, schema)
	}
	return lis
}
func (s *Schemas) Get(name string) *Schema {
	return s.m[name]
}

type Parser struct {
	scanner *bufio.Scanner
	current *Object
}

func NewParser(in io.Reader) *Parser {
	return &Parser{scanner: bufio.NewScanner(in)}
}
func (p *Parser) Current() *Object {
	if p.current == nil {
		return &Object{keys: map[string][]int{}}
	}

	return p.current
}
func (p *Parser) Scan() bool {
	var dom *Object

	lineno := 0
	found := false

	for p.scanner.Scan() {
		line := p.scanner.Text()

		if lineno == 0 && line == "" {
			continue
		}
		if lineno > 0 && line == "" {
			break
		}
		if !found {
			found = true
			dom = &Object{}
			dom.keys = make(map[string][]int)
		}

		lineno += 1

		r, _ := utf8.DecodeRuneInString(line)
		switch r {
		case ' ', '\t', '+':
			if len(dom.attributes) == 0 {
				continue
			}
			last := len(dom.attributes) - 1

			if r == '+' {
				dom.attributes[last].rows = append(
					dom.attributes[last].rows, Value{Lineno: lineno})
			} else {
				sp := strings.SplitN(line, "#", 2)

				v := Value{
					Lineno: lineno,
					Value:  strings.TrimSpace(sp[0]),
				}
				if len(sp) == 2 {
					v.Comment = strings.TrimSpace(sp[1])
				}

				dom.attributes[last].rows = append(
					dom.attributes[last].rows, v)
			}

		default:
			sp := strings.SplitN(line, ":", 2)
			if len(sp) < 2 {
				continue
			}
			attr := &Attribute{Name: strings.TrimSpace(sp[0])}
			sp = strings.SplitN(sp[1], "#", 2)
			v := Value{
				Lineno: lineno,
				Value:  strings.TrimSpace(sp[0]),
			}
			if len(sp) == 2 {
				v.Comment = strings.TrimSpace(sp[1])
			}
			attr.rows = append(attr.rows, v)
			if lis, ok := dom.keys[attr.Name]; ok {
				dom.keys[attr.Name] = append(lis, len(dom.attributes))
			} else {
				dom.keys[attr.Name] = []int{len(dom.attributes)}
			}
			dom.attributes = append(dom.attributes, attr)
		}
	}

	p.current = dom
	return found
}

type SchemaParser struct {
	keys *Set
}

func NewSchemaParser(keys ...string) *SchemaParser {
	return &SchemaParser{keys: NewSet(keys...)}
}
func (p *SchemaParser) ParseSchema(dom *Object) *Schema {
	schema := &Schema{}
	schema.Links = make(map[string][]string)
	schema.spec = make(map[string]Spec)
	schema.specTx = make(map[string][]string)
	schema.Rules = make(map[string]*Set)

	first := true
	for _, row := range dom.Attrs() {
		switch row.Name {
		case "key":
			fields := row.Fields()
			key := fields[0]
			fields = fields[1:]

			if first {
				first = false
				schema.Name = key
				schema.Primary = key
			}

			schema.Rules[key] = NewSet()
			for i, rule := range fields {
				if rule == ">" {
					schema.specTx[key] = fields[i+1:]
					break
				}

				schema.Rules[key].Add(rule)
			}
		}
	}

	for key, rules := range schema.Rules {
		if rules.Has("primary") {
			schema.Primary = key
			rules.Add("oneline", "single", "required")
			rules.Del("multiline", "optional", "recommend", "multiline")
		}
		if !rules.Has("oneline") {
			rules.Add("multiline")
		}
		if !rules.Has("single") {
			rules.Add("multiple")
		}
	}

	return schema
}
func (p *SchemaParser) ParseSpec(lis []string) (Spec, error) {
	spec := make([]SpecRule, len(lis))
	for i, s := range lis {
		if strings.ContainsRune(s, '|') {
			options := strings.Split(s, "|")
			rule := make(SpecRulePipe, len(options))
			for j, o := range options {
				r, err := p.parseSpecRule(o)
				if err != nil {
					return nil, err
				}
				rule[j] = r
			}
			spec[i] = rule
			continue
		}

		r, err := p.parseSpecRule(s)
		if err != nil {
			return nil, err
		}
		spec[i] = r
	}

	return spec, nil
}
func (p *SchemaParser) parseSpecRule(o string) (SpecRule, error) {
	switch {
	case o[0] == '{' && o[len(o)-1] == '}':
		o = o[1 : len(o)-1]

		rule := &SpecRuleEnum{}
		if strings.ContainsRune(o, ':') {
			sp := strings.SplitN(o, ":", 2)
			rule.Name = sp[0]
			o = sp[1]
		}
		rule.Choices = NewSet(strings.Split(o, ",")...)

		return rule, nil
	case o[0] == '[' && o[len(o)-1] == ']':
		o = o[1 : len(o)-1]

		if strings.ContainsRune(o, ':') {
			sp := strings.SplitN(o, ":", 2)
			if !strings.ContainsRune(sp[1], ',') && !p.keys.Has(sp[1]) {
				return &SpecRuleLabel{Name: sp[0], Type: sp[1]}, nil
			}

			choices := strings.Split(sp[1], ",")
			if !p.keys.Has(choices...) {
				return nil, fmt.Errorf("choices %v arn't all known objects", choices)
			}

			return &SpecRuleLookup{Name: sp[0], Choices: choices}, nil
		}

		if strings.ContainsRune(o, ',') {
			return nil, fmt.Errorf("rule %s contains invalid rune ','", o)
		}

		return &SpecRuleLabel{Name: o}, nil
	case o[0] == '\'' && o[len(o)-1] == '\'':
		o = o[1 : len(o)-1]
		rule := SpecRuleConst(o)

		return rule, nil
	case o[0] == '.' && o[len(o)-1] == '.':
		return &SpecRuleText{}, nil
	default:
		return nil, fmt.Errorf("invalid rule: %s", o)
	}
}

type Spec []SpecRule

func (ls Spec) String() string {
	lis := make([]string, len(ls))
	for i, rule := range ls {
		lis[i] = rule.String()
	}
	return strings.Join(lis, " ")
}

type SpecRule interface {
	ApplyArgument(*Arguments, []string) int
	fmt.Stringer
}

var _ SpecRule = (*SpecRuleEnum)(nil)

type SpecRuleEnum struct {
	Name    string
	Choices *Set
}

func (rule *SpecRuleEnum) String() string {
	if rule == nil {
		return ""
	}
	inner := rule.Choices.String()
	if rule.Name != "" {
		inner = rule.Name + ":" + inner
	}

	return "{" + inner + "}"
}
func (rule *SpecRuleEnum) ApplyArgument(args *Arguments, input []string) int {
	if len(input) == 0 || !rule.Choices.Has(input[0]) {
		if rule.Name == "" {
			for _, name := range rule.Choices.Members() {
				args.Set(name, BoolArg(false))
			}
		}

		return 0
	}

	s := input[0]
	if rule.Name != "" {
		args.Set(rule.Name, StringArg(s))
		return 1
	}

	for _, name := range rule.Choices.Members() {
		args.Set(name, BoolArg(name == s))
	}

	return 1
}

var _ SpecRule = (*SpecRuleLabel)(nil)

type SpecRuleLabel struct {
	Name string
	Type string
}

func (rule *SpecRuleLabel) String() string {
	if rule == nil {
		return ""
	}
	inner := rule.Name
	if len(rule.Type) > 0 {
		inner += ":" + rule.Type
	}

	return "[" + inner + "]"
}
func (rule *SpecRuleLabel) ApplyArgument(args *Arguments, input []string) int {
	if len(input) == 0 {
		return 0
	}
	s := input[0]

	switch rule.Type {
	case "str", "":
		args.Set(rule.Name, StringArg(s))
		return 1

	case "int":
		i, err := strconv.Atoi(s)
		if err != nil {
			args.Set(rule.Name, &ErrArg{Err: err, Text: s})
			return 1
		}
		args.Set(rule.Name, IntArg(i))
		return 1

	case "float":
		fl, err := strconv.ParseFloat(s, 64)
		if err != nil {
			args.Set(rule.Name, &ErrArg{Err: err, Text: s})
			return 1
		}
		args.Set(rule.Name, FloatArg(fl))
		return 1

	case "bool":
		b, err := strconv.ParseBool(s)
		if err != nil {
			args.Set(rule.Name, &ErrArg{Err: err, Text: s})
			return 1
		}
		args.Set(rule.Name, BoolArg(b))
		return 1

	case "email":
		n := 1

		if !strings.ContainsRune(s, '@') {
			for i := 1; i < len(input); i++ {
				if strings.ContainsRune(input[i], '@') || input[i][0] == '<' && input[i][len(input[i])] == '>' {
					n = i
					s = strings.Join(input[:i+1], " ")
					break
				}
			}
		}

		a, err := mail.ParseAddress(s)
		if err != nil {
			args.Set(rule.Name, &ErrArg{Err: err, Text: s})
			return n
		}

		args.Set(rule.Name, (*EmailArg)(a))

		return n

	default:
		return 0
	}
}

var _ SpecRule = (*SpecRuleLookup)(nil)

type SpecRuleLookup struct {
	Name    string
	Choices []string
}

func (rule *SpecRuleLookup) String() string {
	if rule == nil {
		return ""
	}
	inner := rule.Name
	if len(rule.Choices) > 0 {
		inner += ":" + strings.Join(rule.Choices, ",")
	}

	return "[" + inner + "]"
}
func (rule *SpecRuleLookup) ApplyArgument(args *Arguments, input []string) int {
	for _, s := range input {
		args.Set(rule.Name, &LookupArg{Value: s, Choices: rule.Choices})
		return 1
	}

	return 0
}

var _ SpecRule = SpecRuleConst("")

type SpecRuleConst string

func (rule SpecRuleConst) String() string {
	return "'" + string(rule) + "'"
}
func (rule SpecRuleConst) ApplyArgument(args *Arguments, input []string) int {
	for i, s := range input {
		if s == string(rule) {
			return i + 1
		}
	}

	return len(input)
}

var _ SpecRule = (*SpecRuleText)(nil)

type SpecRuleText struct{}

func (rule *SpecRuleText) String() string {
	return "..."
}
func (rule *SpecRuleText) ApplyArgument(args *Arguments, input []string) int {
	args.Set("...", StringArg(strings.Join(input, " ")))
	return len(input)
}

var _ SpecRule = SpecRulePipe(nil)

type SpecRulePipe []SpecRule

func (rule SpecRulePipe) String() string {
	lis := make([]string, len(rule))
	for i, r := range rule {
		lis[i] = r.String()
	}
	return strings.Join(lis, "|")
}
func (rule SpecRulePipe) ApplyArgument(args *Arguments, input []string) int {
	l := 0

	for _, r := range rule {
		l += r.ApplyArgument(args, input)
		if l > 0 {
			break
		}
	}

	return l
}

type Arguments struct {
	m map[string]Argument
}

func NewArguments() *Arguments {
	return &Arguments{m: make(map[string]Argument)}
}
func (a *Arguments) Add(name string, arg Argument) *Arguments {
	a.m[name] = arg
	return a
}
func (a *Arguments) String() string {
	lis := make([]string, len(a.m))
	i := 0
	for k, v := range a.m {
		lis[i] = fmt.Sprintf("%s:%v", k, v)
		i++
	}
	sort.Strings(lis)

	return strings.Join(lis, " ")
}
func (a *Arguments) Has(name string) bool {
	_, ok := a.m[name]
	return ok
}
func (a *Arguments) Get(name string) Argument {
	return a.m[name]
}
func (a *Arguments) Set(name string, arg Argument) {
	a.m[name] = arg
}
func (a *Arguments) Keys() []string {
	lis := make([]string, len(a.m))
	i := 0
	for name := range a.m {
		lis[i] = name
		i++
	}
	sort.Strings(lis)
	return lis
}
func (a *Arguments) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.m)
}

type Argument interface {
	isArgument()
	String() string
}

var _ Argument = (*LookupArg)(nil)

type LookupArg struct {
	Value   string
	Choices []string
}

func (a *LookupArg) isArgument() {}
func (a *LookupArg) String() string {
	var b strings.Builder
	for i, c := range a.Choices {
		if i > 0 {
			b.WriteRune('|')
		}
		b.WriteString(c)
		b.WriteRune('/')
		b.WriteString(a.Value)
	}
	return b.String()
}
func (a *LookupArg) Lookups() [][2]string {
	lis := make([][2]string, len(a.Choices))
	for i, c := range a.Choices {
		lis[i] = [2]string{c, a.Value}
	}
	return lis
}

var _ Argument = StringArg("")

type StringArg string

func (s StringArg) isArgument() {}
func (s StringArg) Format(f fmt.State, c rune) {
	switch c {
	case 'v':
		fmt.Fprintf(f, `"%s"`, s.String())
	default:
		fmt.Fprint(f, s.String())
	}
}
func (s StringArg) String() string {
	return string(s)
}

var _ Argument = (*ErrArg)(nil)

type ErrArg struct {
	Err  error
	Text string
}

func (e *ErrArg) isArgument() {}
func (e *ErrArg) String() string {
	return e.Err.Error()
}

var _ Argument = (*EmailArg)(nil)

type EmailArg mail.Address

func (s *EmailArg) isArgument() {}
func (s *EmailArg) String() string {
	if s == nil {
		return ""
	}

	return (*mail.Address)(s).String()
}

type BoolArg bool

func (s BoolArg) isArgument() {}
func (s BoolArg) String() string {
	if bool(s) {
		return "true"
	}
	return "false"
}

var _ Argument = IntArg(0)

type IntArg int64

func (s IntArg) isArgument() {}
func (s IntArg) String() string {
	return strconv.Itoa(int(s))
}

var _ Argument = FloatArg(0.0)

type FloatArg float64

func (s FloatArg) isArgument() {}
func (s FloatArg) String() string {
	return fmt.Sprint(float64(s))
}

var _ Argument = (*Set)(nil)

type Set struct {
	set map[string]struct{}
}

func NewSet(members ...string) *Set {
	s := &Set{set: make(map[string]struct{})}
	for _, member := range members {
		s.Add(member)
	}
	return s
}
func (s *Set) isArgument() {}
func (s *Set) Has(names ...string) bool {
	found := false
	for _, n := range names {
		if _, found = s.set[n]; !found {
			return false
		}
	}

	return found
}
func (s *Set) Add(names ...string) {
	for _, name := range names {
		s.set[name] = struct{}{}
	}
}
func (s *Set) Del(names ...string) {
	for _, name := range names {
		delete(s.set, name)
	}
}
func (s *Set) String() string {
	lis := s.Members()
	return strings.Join(lis, ",")
}
func (s *Set) Members() []string {
	lis := make([]string, len(s.set))
	i := 0
	for name := range s.set {
		lis[i] = name
		i++
	}
	sort.Strings(lis)
	return lis
}
func (s *Set) MarshalJSON() ([]byte, error) {
	lis := s.Members()
	return []byte(`["` + strings.Join(lis, `", "`) + `"]`), nil
}

type nullFS struct{}

func (*nullFS) LoadObject(schema, name string) (*Object, error) {
	return nil, NotFound
}
func (*nullFS) FindObject(search string) ([]*Object, error) {
	return nil, NotFound
}

var NotFound = errors.New("object not found")
