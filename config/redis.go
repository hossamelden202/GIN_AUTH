package config
import("github.com/redis/go-redis/v9"
    "context")
var Rdb *redis.Client
var Ctx=context.Background()
func ConnectRedis(){
Rdb=redis.NewClient(&redis.Options{
	Addr:"localhost:6379",
	Password:"",
	DB:0,
})

err:=Rdb.Ping(Ctx).Err()
if err!=nil{
	panic("something went wrong during connection database")
}
// err2:=Rdb.FlushAll(Ctx).Err()
// if err2!=nil{
// 	panic("something went wrong during connection database")
// }
}