package frontend

interface FrontEnd{
}


func ParseHttp(*FrontEnd,* http.Request) CommonRequest

func Protocol(*FrontEnd)string
