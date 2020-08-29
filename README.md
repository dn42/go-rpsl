# RPSL Spec DRAFT

[![Coverage Report](https://paste.dn42.us/s/cover.registry.go-rpsl.badge)](https://paste.dn42.us/s/cover.registry.go-rpsl.report)

```go
| Python                                               | Go                                                                 |
|:-----------------------------------------------------|:-------------------------------------------------------------------|
| from rpsl import RPSL                                | import rpsl.dn42.us/go-rpsl                                        |
|                                                      |                                                                    |
| rpsl = RPSL.schema_file('schema.txt')                | r, err := rpsl.NewRPSL(rpsl.WithSchemaFile("schema.txt"))          |
| rpsl = RPSL.schema_dir('registry/data/schema/')      | r, err = rpsl.NewRPSL(rpsl.WithSchemaDir("registry/data/schema/")) |
| rpsl = RPSL.rpsl_dir("registry/data/")               | r, err = rpsl.NewRPSL(rpsl.WithRPSLDir("registry/data/"))          |
|                                                      |                                                                    |
| mnt = rpsl.read_file('registry/data/mntner/XUU-MNT') | mnt, err := r.ReadFile("registry/data/mntner/XUU-MNT")             |
| mnt = rpsl.read('mntner', 'XUU-MNT')                 | mnt, err = r.Read("mntner", "XUU-MNT")                             |
|                                                      |                                                                    |
| print(mnt.get('descr'))                              | fmt.Println(mnt.Get("descr").Text())                               |
| print(mnt.get_all('auth'))                           | fmt.Println(mnt.GetAll("auth").Text())                             |
| auth = mnt.get('auth', index=2)                      | auth, _ := mnt.GetN("auth", 2)                                     |
|                                                      |                                                                    |
| print(auth.name)                                     | fmt.Println(auth.Name())                                           |
| print(auth.text)                                     | fmt.Println(auth.Text())                                           |
| print(auth.fields)                                   | fmt.Println(auth.Fields())                                         |
| print(auth['type'])                                  | fmt.Println(auth.Get("type"))                                      |
| print(auth['pubkey'])                                | fmt.Println(auth.Get("pubkey"))                                    |
|                                                      |                                                                    |
| print(mnt.get('xxx', default='missing'))             | fmt.Println(mnt.Get("xxx").Default("missing"))                     |
|                                                      |                                                                    |
| admin = mnt.get('admin-c').fetch('lookup')           | admin, err := mnt.Get("admin-c").Fetch("lookup")                   |
| print(admin.get('nic-hdl'))                          | print(admin.Get('nic-hdl'))                                        |
|                                                      |                                                                    |
| mnt.set('auth', "ssh-rsa AAA...")                    | mnt.Set('auth', "ssh-rsa AAA...")                                  |
| mnt.set('auth', index=2, "ssh-rsa AAA...")           | mnt.SetN('auth', 2, "ssh-rsa AAA...")                              |
| mnt.add('auth', "ssh-rsa AAA...")                    | mnt.Add('auth', "ssh-rsa AAA...")                                  |
| mnt.save()                                           | mnt.Save()                                                         |
|                                                      |                                                                    |
|                                                      |                                                                    |
```
