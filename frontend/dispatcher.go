package frontend


type Dispatcher struct{

	 protocolMapFrontend map [string][FrontEnd]

}

var disptacher Dispatcher 

func RegisterFrontend(*Dispatcher,FrontEnd){
}

func Parse(Dispatcher*,request Http.Request){
}
