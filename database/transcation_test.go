package database

import (
	"github.com/xzwsloser/Go-redis/resp/connection"
	"log"
	"testing"
)

func TestTranscation(t *testing.T) {
	db := NewDatabase(0)
	conn := connection.NewConnection(nil)

	command01 := [][]byte{
		[]byte("SET"),
		[]byte("k111"),
		[]byte("v100000"),
	}

	// WATCH 防止另外一个 客户端修改了需要监听的 key
	command0 := [][]byte{
		[]byte("WATCH"),
		[]byte("k111"),
	}

	command1 := [][]byte{
		[]byte("MULTI"),
	}

	command2 := [][]byte{
		[]byte("SET"),
		[]byte("k111"),
		[]byte("v111"),
	}

	command3 := [][]byte{
		[]byte("SET"),
		[]byte("k222"),
		[]byte("v222"),
	}

	command4 := [][]byte{
		[]byte("GET"),
		[]byte("k222"),
	}

	//command41 := [][]byte{
	//	[]byte("KILL"),
	//}

	command5 := [][]byte{
		[]byte("EXEC"),
	}

	reply := db.Exec(conn, command01)
	log.Println("====== the result of set =====")
	log.Print(string(reply.ToByte()))
	log.Println("==============================")

	reply = db.Exec(conn, command0)
	log.Println("====== the result of watch======")
	log.Print(string(reply.ToByte()))
	log.Println("=================================")

	reply = db.Exec(conn, command1)
	log.Println("====== the result of multi ======")
	log.Print(string(reply.ToByte()))
	log.Println("=================================")

	reply = db.Exec(conn, command2)
	log.Println("===== the result of command2 =====")
	log.Print(string(reply.ToByte()))
	log.Println("==================================")

	reply = db.Exec(conn, command3)
	log.Println("===== the result of command3 =====")
	log.Print(string(reply.ToByte()))
	log.Println("==================================")

	reply = db.Exec(conn, command4)
	log.Println("===== the result of command4 =====")
	log.Print(string(reply.ToByte()))
	log.Println("==================================")

	//reply = db.Exec(conn, command41)
	//log.Println("===== the result of error command =====")
	//log.Print(string(reply.ToByte()))
	//log.Println("==================================")

	reply = db.Exec(conn, command5)
	log.Println("===== the result of exec =====")
	log.Print(string(reply.ToByte()))
	log.Println("==============================")

	reply = db.Exec(conn, command4)
	log.Println("===== the result of command4 =====")
	log.Print(string(reply.ToByte()))
	log.Println("==================================")
}
