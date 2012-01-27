from sj import *
URL = "http://127.0.0.1:2000/rpc"
service = ServerProxy(URL, verbose=True)
print service.call("Responder.Get", "test")
print service.ResponseGet("test")
