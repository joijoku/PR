package shared

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

type Exception interface{}

func Throw(up Exception) {
	panic(up)
}

type Block struct {
	Try     func()
	Catch   func(Exception)
	Finally func()
}

func (tcf Block) Do() {
	if tcf.Finally != nil {

		defer tcf.Finally()
	}
	if tcf.Catch != nil {
		defer func() {
			if r := recover(); r != nil {
				tcf.Catch(r)
			}
		}()
	}
	tcf.Try()
}
