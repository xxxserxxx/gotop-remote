module github.com/xxxserxxx/gotop-remote

go 1.14

require (
	github.com/stretchr/testify v1.4.0
	github.com/xxxserxxx/gotop/v3 v3.5.1
	github.com/xxxserxxx/gotop/v4 v4.0.0-20200423190708-ccbc71755fdb
	github.com/xxxserxxx/opflag v1.0.3
)

replace github.com/xxxserxxx/gotop/v4 => ../gotop
