module pipelined.dev/flac

go 1.13

require (
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mewkiz/flac v1.0.6
	github.com/mewkiz/pkg v0.0.0-20200411195739-f6b5e26764c3 // indirect
	github.com/stretchr/testify v1.4.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	pipelined.dev/pipe v0.8.1
	pipelined.dev/signal v0.7.2
	pipelined.dev/wav v0.4.0
)

replace pipelined.dev/wav => ../wav
