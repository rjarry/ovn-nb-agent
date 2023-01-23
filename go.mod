module github.com/rjarry/ovn-nb-agent

go 1.18

require (
	github.com/akamensky/argparse v1.4.0 // MIT
	github.com/ovn-org/libovsdb v0.6.0 // Apache-2.0
	github.com/vishvananda/netlink v1.2.1-beta.2.0.20230420174744-55c8b9515a01 // Apache-2.0
	github.com/vishvananda/netns v0.0.4 // Apache-2.0
	golang.org/x/sys v0.5.0 // BSD-3-Clause
)

require (
	github.com/cenkalti/backoff/v4 v4.1.1 // indirect; MIT; indirect
	github.com/cenkalti/hub v1.0.1 // indirect; MIT; indirect
	github.com/cenkalti/rpc2 v0.0.0-20210604223624-c1acbc6ec984 // indirect; MIT; indirect
	github.com/davecgh/go-spew v1.1.1 // indirect; ISC; indirect
	github.com/google/uuid v1.3.0 // BSD-3-Clause; indirect
	github.com/kr/pretty v0.3.0 // indirect; MIT; indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect; BSD-3-Clause; indirect
	github.com/stretchr/testify v1.8.2 // indirect; MIT; indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect; BSD-2-Clause; indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect; Apache-2.0, MIT; indirect
)

replace github.com/vishvananda/netlink => github.com/rjarry/netlink v0.0.0-20230524111524-85715a9e7efd
