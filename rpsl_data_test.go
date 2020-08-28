package rpsl_test

import (
	"strings"
	"unicode"
)

const (
	txtAllObjects = txtMetaSchemas + txtSchemas + txtDN42Objects + txtSourisObjects

	txtMetaSchemas = `
        schema: META-TS-SCHEMA
        key:    timestamp required single primary > [date] [time] [tz]

        schema: META-SIG-SCHEMA
        key:    signature required single primary > [signature]

        schema: META-TX-BEGIN-SCHEMA
        key:    transaction-submit-begin required single   primary  > [database] [tx]
        key:    response-auth-type       optional multiple          > [type]
        key:    transaction-confirm-type optional single            > {none,legacy,normal,commit}

        schema: META-TX-END-SCHEMA
        key:    transaction-submit-end required single primary > [database] [tx]

        schema: META-TX-CONFIRM-SCHEMA
        key:    transaction-confirm    required single primary > [database] [tx]

        schema: META-TX-LABEL-SCHEMA
        key:    transaction-label required single primary
        key:    sequence
        key:    timestamp
        key:    integrity
    `

	txtSchemas = txtSchemaSchema + txtOtherSchemas

	txtSchemaSchema = `
        schema:             SCHEMA-SCHEMA
        primary-key:        inetnum  cidr
        primary-key:        inet6num cidr
        primary-key:        role     nic-hdl
        primary-key:        person   nic-hdl
        owners:             mntner
        key:                schema           required single    primary
        key:                owners           optional single    > [schema]
        key:                mnt-by           required multiple  > [lookup:mntner]
        key:                remarks          optional multiple
        key:                source           required single    > [lookup:registry]
        key:                network-owner    optional multiple  > [parent-schema:schema]
        key:                key              required multiple  > [name]
                            {required,optional,recommend,deprecate}
                            {single,multiple} {primary,} '>' ...
        mnt-by:             DN42-MNT
        source:             DN42
        remarks:            # option descriptions
                            Attribute names must match /[a-zA-Z]([a-zA-Z0-9_\-]*[a-zA-Z0-9])?/.
        +
                            schema
                            :    the first field name is the schema
        +
                            required
                            :    object required to have at least one
                            optional
                            :    object not required to have at least one
        +
                            single
                            :    only one of this type allowed
                            multiple
                            :    more than one of this type allowed
        +
                            primary
                            :    use alternate field for schema lookup
                            * only one allowed per schema
                            * does not allow newlines
        +
                            lookup
                            :    schema match to use for related record
        +
                            \> option specs
                            :    defines the option specifications for the key.
                            * must come last in option list
        +
                            [label] string value. A positional string argument required.
                            Text inside brackets represent a label for the string and must match the same rules as attribute names.
        +
                            {enum1|enum2|} enumeration. One option in pipe('|') deliniation is allowed.
                            If there is a trailing pipe it means the enum is optional. Enum values must match the same rules as attribute names.
        +
                            'literal' Literal value. literal text value which must not contain any whitespace or single quotes.
    `

	txtOtherSchemas = `
       schema:             AS-BLOCK-SCHEMA
        key:                as-block   required  single    primary
        key:                descr      optional  single
        key:                policy     required  single    > {policy:open,ask,closed}
        key:                mnt-by     required  multiple  > [lookup:mntner]
        key:                admin-c    optional  multiple  > [lookup:nic-hdl]
        key:                tech-c     optional  multiple  > [lookup:nic-hdl]
        key:                remarks    optional  multiple
        key:                source     required  single    > [lookup:registry]
        mnt-by:             DN42-MNT
        source:             DN42

        schema:             AS-SET-SCHEMA
        key:                as-set      required  single    primary
        key:                descr       optional  single
        key:                mnt-by      required  multiple  > [lookup:mntner]
        key:                members     optional  multiple  > [lookup:aut-num,as-set]
        key:                mbrs-by-ref optional  multiple  > [lookup:mntner]
        key:                admin-c     optional  multiple  > [lookup:nic-hdl]
        key:                tech-c      optional  multiple  > [lookup:nic-hdl]
        key:                remarks     optional  multiple
        key:                source      required  single    > [lookup:registry]
        as-owner:           as-block
        mnt-by:             DN42-MNT
        source:             DN42

        schema:             AUT-NUM-SCHEMA
        key:                aut-num    required  single   primary
        key:                as-name    required  single   index
        key:                descr      optional  single
        key:                mnt-by     required  multiple > [lookup:mntner]
        key:                member-of  optional  multiple > [lookup:as-set,route-set]
        key:                admin-c    optional  multiple > [lookup:nic-hdl]
        key:                tech-c     optional  multiple > [lookup:nic-hdl]
        key:                org        optional  single   > [lookup:organisation]
        key:                import     deprecate multiple
        key:                export     deprecate multiple
        key:                default    deprecate multiple
        key:                mp-peer    deprecate multiple
        key:                mp-group   deprecate multiple
        key:                mp-import  optional  multiple
        key:                mp-export  optional  multiple
        key:                mp-default optional  multiple
        key:                geo-loc    optional  multiple
        key:                remarks    optional  multiple
        key:                source     required  single   > [lookup:registry]
        mnt-by:             DN42-MNT
        source:             DN42

        schema:             DNS-SCHEMA
        key:                dns        required   single    primary
        key:                nserver    required   multiple  > [dns] [addr]
        key:                descr      optional   single
        key:                mnt-by     required   multiple  > [lookup:mntner]
        key:                admin-c    optional   multiple  > [lookup:nic-hdl]
        key:                tech-c     optional   multiple  > [lookup:nic-hdl]
        key:                org        optional   multiple  > [lookup:organisation]
        key:                country    optional   single
        key:                ds-rdata   optional   multiple
        key:                remarks    optional   multiple
        key:                source     required   single    > [lookup:registry]
        mnt-by:             DN42-MNT
        source:             DN42

        schema:             INET6NUM-SCHEMA
        key:                inet6num    required   single
        key:                cidr        required   single    primary
        key:                netname     required   single
        key:                nserver     optional   multiple  > [dns]
        key:                country     optional   multiple
        key:                descr       optional   single
        key:                status      optional   single    > {space:ALLOCATED,ASSIGNED} {use:ANYCAST,}
        key:                policy      optional   single    > {policy:open,closed,ask,reserved}
        key:                admin-c     optional   multiple  > [lookup:nic-hdl]
        key:                tech-c      optional   multiple  > [lookup:nic-hdl]
        key:                zone-c      optional   multiple  > [lookup:nic-hdl]
        key:                ds-rdata    optional   multiple
        key:                mnt-by      optional   multiple  > [lookup:mntner]
        key:                mnt-lower   optional   multiple  > [lookup:mntner]
        key:                mnt-routes  optional   multiple  > [lookup:mntner]
        key:                org         optional   single    > [lookup:organisation]
        key:                remarks     optional   multiple
        key:                source      required   single    > [lookup:registry]
        network-owner:      inet6num
        mnt-by:             DN42-MNT
        source:             DN42

        schema:             INETNUM-SCHEMA
        key:                inetnum     required  single    schema
        key:                cidr        required  single    primary
        key:                netname     required  single
        key:                nserver     optional  multiple  > [dns]
        key:                country     optional  multiple
        key:                descr       optional  single
        key:                status      optional  single    > {space:ALLOCATED,ASSIGNED} {use:ANYCAST,}
        key:                policy      optional  single    > {policy:open,closed,ask,reserved}
        key:                admin-c     optional  multiple  > [lookup:nic-hdl]
        key:                tech-c      optional  multiple  > [lookup:nic-hdl]
        key:                zone-c      optional  multiple  > [lookup:nic-hdl]
        key:                ds-rdata    optional  multiple
        key:                mnt-by      optional  multiple  > [lookup:mntner]
        key:                mnt-lower   optional  multiple  > [lookup:mntner]
        key:                mnt-routes  optional  multiple  > [lookup:mntner]
        key:                org         optional  single    > [lookup:organisation]
        key:                remarks     optional  multiple
        key:                source      required  single    > [lookup:registry]
        network-owner:      inet6num
        network-owner:      inetnum
        mnt-by:             DN42-MNT
        source:             DN42

        schema:             KEY-CERT-SCHEMA
        key:                key-cert     required  single    primary
        key:                method       required  single    > {type:PGP,X509,MTN,SSH}
        key:                owner        required  multiple  > [email]
        key:                fingerpr     required  single
        key:                certif       required  multiple
        key:                org          optional  multiple  > [lookup:organisation]
        key:                remarks      optional  multiple
        key:                admin-c      optional  multiple  > [lookup:nic-hdl]
        key:                tech-c       optional  multiple  > [lookup:nic-hdl]
        key:                mnt-by       required  multiple  > [lookup:mntner]
        key:                source       required  single    > [lookup:registry]
        mnt-by:             DN42-MNT
        source:             DN42

        schema:             MNTNER-SCHEMA
        key:                mntner         required  single    primary
        key:                descr          optional  single
        key:                mnt-by         required  multiple  > [lookup:mntner]
        key:                admin-c        optional  multiple  > [lookup:nic-hdl]
        key:                tech-c         optional  multiple  > [lookup:nic-hdl]
        key:                auth           optional  multiple  > {type:ssh-rsa,ssh-ed25519}|[lookup:key-cert] [data]
        key:                org            optional  multiple  > [lookup:organisation]
        key:                abuse-mailbox  optional  single    > [email]
        key:                remarks        optional  multiple
        key:                source         required  single    > [lookup:registry]
        mnt-by:             DN42-MNT
        source:             DN42

        schema:             ORGANISATION-SCHEMA
        key:                organisation   required  single    primary
        key:                org-name       required  single    index
        key:                descr          optional  single
        key:                admin-c        optional  multiple  > [lookup:nic-hdl]
        key:                tech-c         optional  multiple  > [lookup:nic-hdl]
        key:                abuse-c        optional  multiple  > [lookup:nic-hdl]
        key:                mnt-by         required  multiple  > [lookup:mntner]
        key:                mnt-ref        optional  multiple  > [lookup:mntner]
        key:                phone          optional  multiple
        key:                fax-no         optional  multiple
        key:                www            optional  multiple
        key:                abuse-mailbox  optional  multiple  > [email]
        key:                e-mail         optional  multiple  > [email]
        key:                geoloc         optional  multiple  > [lat-c] [long-c] [name]
        key:                language       optional  multiple
        key:                remarks        optional  multiple
        key:                address        optional  multiple
        key:                source         required  single    > [lookup:registry]
        mnt-by:             DN42-MNT
        source:             DN42

        schema:             PERSON-SCHEMA
        key:                person          required   single    schema
        key:                nic-hdl         required   single    primary
        key:                mnt-by          required   multiple  > [lookup:mntner]
        key:                org             optional   multiple  > [lookup:organisation]
        key:                nick            optional   multiple
        key:                pgp-fingerprint optional   multiple
        key:                www             optional   multiple
        key:                e-mail          optional   multiple  > [email]
        key:                contact         optional   multiple  > [email]
        key:                abuse-mailbox   optional   multiple  > [email]
        key:                phone           optional   multiple
        key:                fax-no          optional   multiple
        key:                address         optional   multiple
        key:                remarks         optional   multiple
        key:                source          required   single    > [lookup:registry]
        mnt-by:             DN42-MNT
        source:             DN42

        schema:             REGISTRY-SCHEMA
        key:                registry   required  single    primary
        key:                url        required  multiple
        key:                descr      optional  multiple
        key:                whois      optional  single
        key:                mnt-by     required  multiple  > [lookup:mntner]
        key:                admin-c    optional  multiple  > [lookup:nic-hdl]
        key:                tech-c     optional  multiple  > [lookup:nic-hdl]
        key:                source     required  single    > [lookup:registry]
        mnt-by:             DN42-MNT
        source:             DN42

        schema:             ROLE-SCHEMA
        key:                role            required   single    schema
        key:                nic-hdl         required   single    primary
        key:                mnt-by          required   multiple  > [lookup:mntner]
        key:                org             optional   multiple  > [lookup:organisation]
        key:                admin-c         optional   multiple  > [lookup:person]
        key:                tech-c          optional   multiple  > [lookup:person]
        key:                abuse-c         optional   multiple  > [lookup:person]
        key:                abuse-mailbox   optional   multiple  > [email]
        key:                descr           optional   single
        key:                remarks         optional   multiple
        key:                source          required   single    > [lookup:registry]
        mnt-by:             DN42-MNT
        source:             DN42

        schema:             ROUTE-SCHEMA
        key:                route       required  single    primary
        key:                mnt-by      required  multiple  > [lookup:mntner]
        key:                origin      required  multiple  > [lookup:aut-num]
        key:                member-of   optional  multiple  > [lookup:route-set]
        key:                admin-c     optional  multiple  > [lookup:nic-hdl]
        key:                tech-c      optional  multiple  > [lookup:nic-hdl]
        key:                descr       optional  single
        key:                remarks     optional  multiple
        key:                source      required  single    > [lookup:registry]
        key:                pingable    optional  multiple
        key:                max-length  optional  single
        network-owner:      inetnum
        mnt-by:             DN42-MNT
        source:             DN42

        schema:             ROUTE-SET-SCHEMA
        key:                route-set    required  single    primary
        key:                descr        optional  single
        key:                mnt-by       required  multiple  > [lookup:mntner]
        key:                members      deprecate multiple
        key:                mp-members   optional  multiple
        key:                mbrs-by-ref  optional  multiple  > [lookup:mntner]
        key:                admin-c      optional  multiple  > [lookup:nic-hdl]
        key:                tech-c       optional  multiple  > [lookup:nic-hdl]
        key:                remarks      optional  multiple
        key:                source       required  single    > [lookup:registry]
        mnt-by:             DN42-MNT
        source:             DN42

        schema:             ROUTE6-SCHEMA
        key:                route6      required  single    primary
        key:                mnt-by      required  multiple  > [lookup:mntner]
        key:                origin      required  multiple  > [lookup:aut-num]
        key:                member-of   optional  multiple  > [lookup:route-set]
        key:                admin-c     optional  multiple  > [lookup:nic-hdl]
        key:                tech-c      optional  multiple  > [lookup:nic-hdl]
        key:                descr       optional  multiple
        key:                remarks     optional  multiple
        key:                source      required  single    > [lookup:registry]
        key:                pingable    optional  multiple
        key:                max-length  optional  single
        network-owner:      inet6num
        mnt-by:             DN42-MNT
        source:             DN42

        schema:             TINC-KEY-SCHEMA
        key:                tinc-key      required  single    primary
        key:                tinc-host     required  single
        key:                tinc-file     required  single
        key:                descr         optional  single
        key:                remarks       optional  multiple
        key:                compression   optional  single
        key:                subnet        optional  multiple
        key:                tinc-address  optional  single
        key:                port          optional  single
        key:                admin-c       optional  multiple  > [lookup:nic-hdl]
        key:                tech-c        optional  multiple  > [lookup:nic-hdl]
        key:                mnt-by        required  multiple  > [lookup:mntner]
        key:                source        required  single    > [lookup:registry]
        mnt-by:             DN42-MNT
        source:             DN42
    `

	txtDN42Objects = txtDN42Mntner + txtDN42Inetnum + txtDN42Registry

	txtDN42Mntner = `
        mntner:             DN42-MNT
        descr:              mntner for owning objects in the name of whole dn42.
        mnt-by:             DN42-MNT
        source:             DN42
    `

	txtDN42Inetnum = `
        inetnum:            0.0.0.0 - 255.255.255.255
        cidr:               0.0.0.0/0
        netname:            NET-BLK0-DN42
        policy:             open
        descr:              * The entire IPv4 address space
        mnt-by:             DN42-MNT
        status:             ALLOCATED
        source:             DN42
    `

	txtDN42Registry = `
        registry:           DN42
        url:                https://git.dn42.us/dn42/registry
        mnt-by:             DN42-MNT
        source:             DN42
    `
	txtSourisObjects = txtRoleObject + txtPersonObject + txtMnterObject + txtInetnumObject

	txtRoleObject = `
        role:               Souris Organization Role
        abuse-mailbox:      abuse@sour.is
        admin-c:            XUU-DN42
        tech-c:             XUU-DN42
        nic-hdl:            SOURIS-DN42
        mnt-by:             XUU-MNT
        source:             DN42
    `
	txtPersonObject = `
        person:             Xuu
        contact:            xmpp:xuu@xmpp.dn42
        contact:            mail:xuu@dn42.us
        remarks:            test
                            foo
        +
                            bar
        pgp-fingerprint:    20AE2F310A74EA7CEC3AE69F8B3B0604F164E04F
        nic-hdl:            XUU-DN42
        mnt-by:             XUU-MNT
        source:             DN42
    `
	txtMnterObject = `
        mntner:             XUU-MNT
        descr:              Xuu Maintenance Object
        admin-c:            SOURIS-DN42
        tech-c:             SOURIS-DN42
        mnt-by:             XUU-MNT
        auth:               PGP-LASKJd
        auth:               ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAINpgNnxR4KvmhE9MNF4vNUhtHS8SlKMqdgX43BMvVOhL
        auth:               ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAv6dr5ENW8MYdxf2wot2IDoHfqiYKbT800+STp1qOSKP8LHz7Cx0WfzAo29/sIQCd88hGKppt8XMsu5V0zqXP7+DZaIgrq4+Zt4OpaOzBvKgf7bVdJ+ygh42SDzyMz70wNACuB2saFvlFejMKY1E6R/wkBDQYbURhWMjMgtjc/jjIAGFM7BlwZsF2dkOBRRhzH6Qc0HQAXjEcbtlo5gsu7aSWI0vEDMA9a9F1Ql5SC9sT/kGCJUAMZNtt8i5MQ+iQ3FY3ur3wUcaimbfjAjqjvOy6Lgmm7B6zg91VR98htdbhgvDk3LdxFSWMl1XnkE3cku231NLTrTpwxs4smBFR/Q==
        source:             DN42
    `
	txtInetnumObject = `
        inetnum:            172.21.64.0 - 172.21.64.7
        cidr:               172.21.64.0/29
        netname:            XUU-TEST-NET
        descr:              Xuu TestNet
        country:            US
        admin-c:            SOURIS-DN42
        tech-c:             SOURIS-DN42
        mnt-by:             XUU-MNT
        nserver:            lavana.sjc.xuu.dn42
        nserver:            kapha.mtr.xuu.dn42
        nserver:            rishi.bre.xuu.dn42
        status:             ALLOCATED
        remarks:            This is a transfernet.
        source:             DN42
    `

	txtFooObject = `
        empty:              value
        foo:                bar
        other:              one
                            two # comment two
                            three # comment three
        none:               `

	txtFooObject2 = `
        empty:                        value
        foo:                          bar
        other:                        one
                                      two # comment two
                                      three # comment three
        none:                         ` + `
        something-very-long-past-19:  `

	txtFooSchema = `
        schema:             empty
        key:                empty required single
        key:                foo required single primary
    `
)

func countLeadingSpace(line string) int {
	i := 0
	for _, runeValue := range line {
		if runeValue == ' ' || runeValue == '\t' {
			i++
		} else if unicode.IsSpace(runeValue) {
			i++
		} else {
			break
		}
	}
	return i
}

func cleanDoc(in string) string {
	pad := 0
	sp := strings.Split(in, "\n")
	out := make([]string, 0, len(sp))

	for _, line := range sp {
		if len(line) == 0 && len(out) == 0 {
			continue
		}
		if len(line) == 0 {
			out = append(out, line)
			continue
		}

		hasPad := countLeadingSpace(line)
		if pad == 0 && hasPad > 0 {
			pad = hasPad
		}

		start := hasPad
		if start > pad {
			start = pad
		}
		out = append(out, line[start:])
	}

	for i := len(out) - 1; i >= 0; i-- {
		if len(out[i]) > 0 {
			out = out[:i+1]
			break
		}
	}

	return strings.Join(out, "\n")
}
