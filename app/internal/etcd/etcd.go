package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"
)

var Client *clientv3.Client

// NewEtcDClient instancia o singleton para acesso às funcionalidades do etcd
func NewEtcDClient() {

	var err error
	Client, err = clientv3.New(clientv3.Config{Endpoints: []string{"localhost:2379"}})
	if err != nil {
		panic(err)
	}
	//defer cli.Close()

}
