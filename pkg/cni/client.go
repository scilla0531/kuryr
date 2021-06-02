package cni

const KuryrCNIVersion = "1.0.0-beta.1"

// To allow for testing with a fake client.
//var withClient = rpcClient
//
//func rpcClient(f func(client cnipb.CniClient) error) error {
//	conn, err := grpc.Dial(
//		KuryrCNISocketAddr,
//		grpc.WithInsecure(),
//		grpc.WithContextDialer(func(ctx context.Context, addr string) (conn net.Conn, e error) {
//			return util.DialLocalSocket(addr)
//		}),
//	)
//	if err != nil {
//		return err
//	}
//	defer conn.Close()
//	return f(cnipb.NewCniClient(conn))
//}