module github.com/pipelined/flac

go 1.12

require (
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mewkiz/flac v1.0.5-0.20190222151344-f296b7aa7930
	github.com/mewkiz/pkg v0.0.0-20190919212034-518ade7978e2 // indirect
	github.com/pipelined/signal v0.3.1-0.20191010174805-e6a17bf2be3a
	github.com/pipelined/wav v0.2.0
	github.com/stretchr/testify v1.4.0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace (
	github.com/pipelined/signal => ../signal
	github.com/pipelined/wav => ../wav
)
