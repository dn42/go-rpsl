package rpsl_test

import (
	"encoding/json"
	"net/mail"
	"strings"
	"testing"

	"github.com/matryer/is"
	"rpsl.dn42.us/go-rpsl"
)

func TestParse(t *testing.T) {
	is := is.New(t)

	s := cleanDoc(txtPersonObject)

	dom := rpsl.ParseObject(s)
	is.True(dom != nil)

	is.Equal(strings.TrimSpace(dom.String()), strings.TrimSpace(s))
	attr := dom.Get("person")
	is.True(attr != nil)
	if attr != nil {
		is.Equal(attr.Name, "person")
		is.Equal(attr.Text(), "Xuu")
	}
	is.Equal(dom.Get("contact").Text(), "xmpp:xuu@xmpp.dn42")
	is.Equal(dom.GetN("contact", 1).Text(), "mail:xuu@dn42.us")
	is.Equal(dom.GetN("contact", 2).Text(), "")

	is.Equal(dom.GetAll("contact").Text(), cleanDoc(`xmpp:xuu@xmpp.dn42
        mail:xuu@dn42.us`))

	is.True(dom.Get("mntner").Text() == "")

	is.Equal(dom.Get("mntner").Default("default"), "default")
	is.Equal(dom.GetAll("mntner").Default("default"), "default")

	is.Equal(dom.Schema(), "person")

	is.Equal(dom.GetAll("contact").Fields(), []string{"xmpp:xuu@xmpp.dn42", "mail:xuu@dn42.us"})
	is.Equal(dom.Get("remarks").Fields(), []string{"test", "foo", "bar"})

	dom.Set("contact", "xmpp:Xuu@xmpp.dn42")
	dom.SetN("contact", 1, "mail:Xuu@dn42.us")

	dom.Add("contact", "mail:me@sour.is")
	is.Equal(dom.GetN("contact", 2).Text(), "mail:me@sour.is")

	dom.SetN("contact", -1, "mail:jon@xuu.cc")

	is.Equal(dom.GetAll("contact").Text(), cleanDoc(`xmpp:Xuu@xmpp.dn42
        mail:Xuu@dn42.us
        mail:me@sour.is
        mail:jon@xuu.cc`))

	dom.SetN("remarks", -1, "multi", "line", "remarks")
	is.Equal(dom.GetN("remarks", 1).Text(), "multi\nline\nremarks")

	var empty *rpsl.Object
	is.True(empty == nil)
	is.Equal(empty.Name(), "")
	is.Equal(empty.Schema(), "")
	is.Equal(empty.Primary(), "")

	empty = rpsl.ParseObject("")
	is.True(empty != nil)
	is.Equal(empty.Name(), "")
	is.Equal(empty.Schema(), "")
	is.Equal(empty.Primary(), "")

	empty.Add("empty")
	is.True(empty != nil)
	is.Equal(empty.Name(), "")
	is.Equal(empty.Schema(), "empty")
	is.Equal(empty.Primary(), "empty")

	empty.Set("empty", "value")
	is.Equal(empty.Name(), "value")
	is.Equal(empty.Schema(), "empty")
	is.Equal(empty.Primary(), "empty")
	is.Equal(empty.Get("foo").Default("baz"), "baz")
	is.Equal(empty.GetAll("foo").Default("baz"), "baz")

	empty.Add("foo", "bar")
	empty.Add("other", "one", "two # comment two", "three  #comment three ")
	empty.Add("none")
	empty.Add("something-very-long-past-19")

	is.Equal(empty.Get("foo").Default("baz"), "bar")
	is.Equal(empty.GetAll("foo").Default("baz"), "bar")

	emptyString := cleanDoc(txtFooObject2)

	is.Equal(empty.String(), emptyString)

	lis := rpsl.ListObject{
		rpsl.ParseObject(cleanDoc(txtFooSchema)),
	}

	is.Equal(lis.String(), cleanDoc(txtFooSchema))

	es, err := rpsl.ParseSchemas(lis)

	is.NoErr(err)
	is.Equal(len(es.Items()), 1)
	is.True(es.Get("empty") != nil)
	is.Equal(es.Get("empty").String(), cleanDoc(`
        schema: empty
        primary: foo
        empty: multiline,required,single
        foo: oneline,primary,required,single`))

	es.Apply(empty)

	is.Equal(empty.Name(), "bar")
	is.Equal(empty.Schema(), "empty")
	is.Equal(empty.Primary(), "foo")

	empty.Delete("something-very-long-past-19")

	is.Equal(empty.String(), cleanDoc(txtFooObject))
}

func TestParseMulti(t *testing.T) {
	is := is.New(t)

	s := cleanDoc(txtAllObjects)

	lis := rpsl.ParseAll(strings.NewReader(s))
	is.Equal(len(lis), 30)

	schemas, err := rpsl.ParseSchemas(lis)
	is.NoErr(err)

	schemas.Apply(lis...)

	domByName := make(map[string]*rpsl.Object, len(lis))

	for _, dom := range lis {
		domByName[dom.Name()] = dom

		switch dom.Schema() {
		case "inetnum":
			is.Equal(dom.Primary(), "cidr")
		case "role", "person":
			is.Equal(dom.Primary(), "nic-hdl")
		default:
			is.Equal(dom.Primary(), dom.Schema())
		}
	}

	m := domByName["XUU-MNT"]
	all := m.GetAll("auth")
	for _, o := range all {
		args := o.Args()
		if args.Has("lookup") {
			switch a := args.Get("lookup").(type) {
			case *rpsl.LookupArg:
				is.True(a != nil)
				if a != nil {
					is.Equal(a.Value, "PGP-LASKJd")
					is.Equal(a.Choices, []string{"key-cert"})
				}

			default:
				t.Logf("wrong type: %T %#v\n", a, a)
				is.Fail()
			}
			is.Equal(args.Get("lookup").String(), "key-cert/PGP-LASKJd")
		} else {
			is.True(args.Has("data"))
			is.True(args.Has("type"))

			t, ok := args.Get("type").(rpsl.StringArg)
			is.True(ok)
			is.True(t != "")
		}
	}

	admin := m.Get("admin-c")
	is.True(admin != nil)
	is.True(admin.Args() != nil)
	is.True(admin.Args().Get("lookup") != nil)
	switch a := admin.Args().Get("lookup").(type) {
	case *rpsl.LookupArg:
		is.True(a != nil)
		if a != nil {
			is.Equal(a.Value, "SOURIS-DN42")
			is.Equal(a.Choices, []string{"nic-hdl"})
		}
	default:
		t.Logf("wrong type: %T %#v\n", a, a)
		is.Fail()
	}

	m = domByName["SCHEMA-SCHEMA"]
	remarks := m.Get("remarks")
	is.True(remarks.Args().Has("..."))

	// t.Log("SPEC", schemas.Get("schema").Spec("remarks"))
	// t.Log("COMMENT", remarks.Comment(), "CMT")
	// t.Log("RAW", remarks.Raw(), "RAW")

	var out strings.Builder
	enc := json.NewEncoder(&out)
	enc.SetIndent("", "    ")
	if err := enc.Encode(m); err != nil {
		panic(err)
	}

	// fmt.Printf("JSON %s\n", out.String())

	key := m.Get("key").Args()
	is.True(key.Has("required"))
	is.True(key.Has("single"))
	is.True(key.Has("primary"))
	is.True(key.Has("name"))
	is.Equal(key.Get("name").String(), `schema`)

	net := domByName["0.0.0.0/0"]
	is.Equal(net.Get("status").Args().String(), `space:"ALLOCATED"`)
}

func TestMarshaler(t *testing.T) {
	is := is.New(t)
	b, err := json.Marshal([]rpsl.Argument{
		rpsl.BoolArg(true),
		rpsl.IntArg(123),
		rpsl.FloatArg(1.23),
		rpsl.StringArg("string"),
		rpsl.NewSet("one", "two", "three"),
	})
	is.NoErr(err)
	is.Equal(string(b), `[true,123,1.23,"string",["one","three","two"]]`)
}

func TestSet(t *testing.T) {
	is := is.New(t)

	s := rpsl.NewSet("one", "two", "three")

	is.True(s.Has("one"))
	is.True(s.Has("two"))
	is.True(s.Has("three"))

	is.True(!s.Has("four"))
	s.Add("four")
	is.True(s.Has("four"))
	s.Del("four")
	is.True(!s.Has("four"))
	s.Del("four")
	is.True(!s.Has("four"))

	is.Equal(s.String(), "one,three,two")

	b, err := s.MarshalJSON()
	is.NoErr(err)
	is.Equal(string(b), `["one", "three", "two"]`)
}

func TestAttribute(t *testing.T) {
	is := is.New(t)

	tt := cleanDoc(`
        foo:    bar # comment
                bin
                baz
        `)

	o := rpsl.ParseObject(tt)
	lis := o.Attrs()
	is.Equal(len(lis), 1)

	attr := o.Get("foo")
	is.True(attr != nil)
	if attr != nil {
		is.Equal(attr.Name, "foo")

		lines := attr.Lines()
		is.Equal(len(lines), 3)
		is.Equal(lines, []string{"bar", "bin", "baz"})

		is.Equal(attr.Text(), "bar\nbin\nbaz")

		is.Equal(attr.Comment(), "comment")

		is.Equal(attr.Fields(), []string{"bar", "bin", "baz"})

		is.Equal(attr.Raw(), "bar # comment\nbin\nbaz")
	}

	attr = o.Get("missing")
	is.True(attr == nil)
	is.Equal(attr.Default("default"), "default")

	schema, err := rpsl.ParseSchemas(
		rpsl.ListObject{
			rpsl.ParseObject(cleanDoc(`
                                schema: foo
                                key:    foo required single primary > [one] [two]
                        `)),
		},
	)
	is.NoErr(err)

	schema.Apply(o)

	args := o.Get("foo").Args()
	is.True(args != nil)
	is.True(args.Has("one"))
	is.Equal(args.Get("one").String(), "bar")
	is.True(args.Has("two"))
	is.Equal(args.Get("two").String(), "bin")
	is.True(args.Has("..."))
	is.Equal(args.Get("...").String(), "baz")

	b, err := json.Marshal(o)

	is.NoErr(err)
	is.Equal(string(b), `[["foo",{"...":"baz","one":"bar","two":"bin"}]]`)
}

func TestValue(t *testing.T) {
	is := is.New(t)

	v := rpsl.Value{Value: "value", Comment: "comment"}
	is.Equal(v.String(), "value # comment")

	v = rpsl.Value{Value: "value"}
	is.Equal(v.String(), "value")

	v = rpsl.Value{Comment: "comment"}
	is.Equal(v.String(), "# comment")
}

func TestSpecRule(t *testing.T) {
	is := is.New(t)

	tests := []struct {
		rule    rpsl.SpecRule
		ruleStr string
		in      []string
		out     *rpsl.Arguments
		count   int
		argStr  string
	}{
		// SpecRuleEnum
		{
			&rpsl.SpecRuleEnum{Name: "policy", Choices: rpsl.NewSet("open", "closed", "ask", "reserved")},
			"{policy:ask,closed,open,reserved}",
			[]string{"FOO"},
			rpsl.NewArguments(),
			0,
			"",
		},

		{
			&rpsl.SpecRuleEnum{Name: "", Choices: rpsl.NewSet("open", "closed", "ask", "reserved")},
			"{ask,closed,open,reserved}",
			[]string{"FOO"},
			rpsl.NewArguments().
				Add("ask", rpsl.BoolArg(false)).
				Add("open", rpsl.BoolArg(false)).
				Add("closed", rpsl.BoolArg(false)).
				Add("reserved", rpsl.BoolArg(false)),
			0,
			`ask:false closed:false open:false reserved:false`,
		},

		{
			&rpsl.SpecRuleEnum{Name: "policy", Choices: rpsl.NewSet("open", "closed", "ask", "reserved")},
			"{policy:ask,closed,open,reserved}",
			[]string{"open", "FOO"},
			rpsl.NewArguments().Add("policy", rpsl.StringArg("open")),
			1,
			`policy:"open"`,
		},

		{
			&rpsl.SpecRuleEnum{Name: "", Choices: rpsl.NewSet("open", "closed", "ask", "reserved")},
			"{ask,closed,open,reserved}",
			[]string{"ask", "FOO"},
			rpsl.NewArguments().
				Add("ask", rpsl.BoolArg(true)).
				Add("open", rpsl.BoolArg(false)).
				Add("closed", rpsl.BoolArg(false)).
				Add("reserved", rpsl.BoolArg(false)),
			1,
			"ask:true closed:false open:false reserved:false",
		},

		// SpecRuleLabel
		{
			&rpsl.SpecRuleLabel{Name: "data", Type: "str"},
			"[data:str]",
			[]string{"FOO", "BAR"},
			rpsl.NewArguments().Add("data", rpsl.StringArg("FOO")),
			1,
			`data:"FOO"`,
		},

		{
			&rpsl.SpecRuleLabel{Name: "data", Type: "str"},
			"[data:str]",
			[]string{},
			rpsl.NewArguments(),
			0,
			"",
		},

		{
			&rpsl.SpecRuleLabel{Name: "number", Type: "int"},
			"[number:int]",
			[]string{"123", "BAR"},
			rpsl.NewArguments().Add("number", rpsl.IntArg(123)),
			1,
			`number:123`,
		},

		{
			&rpsl.SpecRuleLabel{Name: "number", Type: "float"},
			"[number:float]",
			[]string{"1.23", "BAR"},
			rpsl.NewArguments().Add("number", rpsl.FloatArg(1.23)),
			1,
			`number:1.23`,
		},

		{
			&rpsl.SpecRuleLabel{Name: "bool", Type: "bool"},
			"[bool:bool]",
			[]string{"true", "BAR"},
			rpsl.NewArguments().Add("bool", rpsl.BoolArg(true)),
			1,
			`bool:true`,
		},

		{
			&rpsl.SpecRuleLabel{Name: "admin-c", Type: "email"},
			"[admin-c:email]",
			[]string{"me@sour.is", "BAR"},
			rpsl.NewArguments().Add("admin-c", (*rpsl.EmailArg)(&mail.Address{Address: "me@sour.is"})),
			1,
			`admin-c:<me@sour.is>`,
		},

		{
			&rpsl.SpecRuleLabel{Name: "admin-c", Type: "email"},
			"[admin-c:email]",
			[]string{"Xuu", "(dn42)", "<me@sour.is>", "BAR"},
			rpsl.NewArguments().Add("admin-c", (*rpsl.EmailArg)(&mail.Address{Address: "me@sour.is", Name: "Xuu (dn42)"})),
			2,
			`admin-c:"Xuu (dn42)" <me@sour.is>`,
		},

		// SpecRuleLookup
		{
			&rpsl.SpecRuleLookup{Name: "mntner", Choices: []string{"nic-hdl"}},
			"[mntner:nic-hdl]",
			[]string{"XUU-MNT", "XXX"},
			rpsl.NewArguments().Add("mntner", &rpsl.LookupArg{"XUU-MNT", []string{"nic-hdl"}}),
			1,
			"mntner:nic-hdl/XUU-MNT",
		},

		{
			&rpsl.SpecRuleLookup{Name: "member", Choices: []string{"aut-num", "as-set"}},
			"[member:aut-num,as-set]",
			[]string{"AS12345", "XXX"},
			rpsl.NewArguments().Add("member", &rpsl.LookupArg{"AS12345", []string{"aut-num", "as-set"}}),
			1,
			"member:aut-num/AS12345|as-set/AS12345",
		},

		{
			&rpsl.SpecRuleLookup{Name: "member", Choices: []string{"aut-num", "as-set"}},
			"[member:aut-num,as-set]",
			[]string{},
			rpsl.NewArguments(),
			0,
			"",
		},

		// SpecRuleConst
		{
			rpsl.SpecRuleConst(">"),
			"'>'",
			[]string{">", "XXX"},
			rpsl.NewArguments(),
			1,
			"",
		},

		{
			rpsl.SpecRuleConst(">"),
			"'>'",
			[]string{"XXX", ">", "XXX"},
			rpsl.NewArguments(),
			2,
			"",
		},

		{
			rpsl.SpecRuleConst(">"),
			"'>'",
			[]string{"XXX", "XXX", "XXX"},
			rpsl.NewArguments(),
			3,
			"",
		},

		// SpecRuleText
		{
			&rpsl.SpecRuleText{},
			"...",
			[]string{"XXX", "XXX", "XXX"},
			rpsl.NewArguments().Add("...", rpsl.StringArg("XXX XXX XXX")),
			3,
			`...:"XXX XXX XXX"`,
		},

		// SpecRulePipe
		{
			rpsl.SpecRulePipe{
				&rpsl.SpecRuleEnum{Name: "type", Choices: rpsl.NewSet("ssh-rsa", "ssh-ed25519")},
				&rpsl.SpecRuleLookup{Name: "lookup", Choices: []string{"key-cert"}},
			},
			"{type:ssh-ed25519,ssh-rsa}|[lookup:key-cert]",
			[]string{"ssh-rsa", "XXX"},
			rpsl.NewArguments().Add("type", rpsl.StringArg("ssh-rsa")),
			1,
			`type:"ssh-rsa"`,
		},

		{
			rpsl.SpecRulePipe{
				&rpsl.SpecRuleEnum{Name: "type", Choices: rpsl.NewSet("ssh-rsa", "ssh-ed25519")},
				&rpsl.SpecRuleLookup{Name: "lookup", Choices: []string{"key-cert"}},
			},
			"{type:ssh-ed25519,ssh-rsa}|[lookup:key-cert]",
			[]string{"PGP-ASDFASDF", "XXX"},
			rpsl.NewArguments().Add("lookup", &rpsl.LookupArg{"PGP-ASDFASDF", []string{"key-cert"}}),
			1,
			"lookup:key-cert/PGP-ASDFASDF",
		},
	}

	sp := rpsl.NewSchemaParser("key-cert", "aut-num", "as-set", "nic-hdl")

	for i, tt := range tests {
		_ = i
		// t.Logf("TEST %d, %s :: %s", i, tt.ruleStr, tt.argStr)
		spec, err := sp.ParseSpec([]string{tt.ruleStr})
		is.NoErr(err)
		is.Equal(len(spec), 1)
		is.Equal(spec[0], tt.rule)
		is.Equal(spec.String(), tt.ruleStr)

		arg := rpsl.NewArguments()
		n := tt.rule.ApplyArgument(arg, tt.in)

		is.Equal(tt.rule.String(), tt.ruleStr)
		is.Equal(n, tt.count)
		is.Equal(arg.String(), tt.argStr)

		for _, name := range tt.out.Keys() {
			//	t.Log(name, arg.Get(name))

			is.True(arg.Has(name))
			is.Equal(arg.Get(name), tt.out.Get(name))
		}

		for _, name := range arg.Keys() {
			//	t.Log(name, arg.Get(name))

			is.True(tt.out.Has(name))
			is.Equal(arg.Get(name), tt.out.Get(name))
		}
	}
}
