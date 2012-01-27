import simplejsonrpc as jsonrpc

server = jsonrpc.Server("http://localhost:2000/rpc")
print dir(server)
#print server.Get("test")
