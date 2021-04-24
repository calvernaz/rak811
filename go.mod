module github.com/calvernaz/rak811

go 1.14

require (
	periph.io/x/conn/v3 v3.6.7
	periph.io/x/host/v3 v3.6.7
)

replace periph.io/x/host/v3 v3.6.7 => github.com/calvernaz/host/v3 v3.6.12
