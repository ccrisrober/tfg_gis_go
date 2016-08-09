// Copyright (c) 2015, maldicion069 (Cristian Rodríguez) <ccrisrober@gmail.con>
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.package com.example

// main project main.go
package main

import "fmt"
import "strings"
import "strconv"
import "github.com/jmoiron/jsonq"
import "encoding/json"
import "time"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"
import "math/rand"
import "log"

func sendFight(db *sql.DB, emisor_id int, jq *jsonq.JsonQuery, server *server) {
	// Save die roll value from emisor_id
	stmt, e := db.Prepare("UPDATE `users` SET `rollDice`=" + strconv.Itoa(RandomValue(1, 6)) + "; WHERE `port`=" + strconv.Itoa(emisor_id) + ";")
	if e != nil {
		panic(e.Error())
	}
	defer stmt.Close()
	stmt.Exec()

	// Save die roll value from receiver_id
	receiver_id, _ := jq.Int("Id_enemy")
	stmte, ee := db.Prepare("UPDATE `users` SET `rollDice`=" + strconv.Itoa(RandomValue(1, 6)) + "; WHERE `port`=" + strconv.Itoa(receiver_id) + ";")
	if ee != nil {
		panic(ee.Error())
	}
	defer stmte.Close()
	stmte.Exec()
	msgOthers := "{\"Action\":\"hide\",\"Ids\":[" + strconv.Itoa(receiver_id) + "," + strconv.Itoa(emisor_id) + "]}"
	go func(emisor int, receiver int, msgOthers string) {
		for index, element := range server.clients {
			if index != emisor {
				if element.port == receiver {
					element.Send("{\"Action\":\"fight\",\"Id_enemy\":" + strconv.Itoa(emisor_id) + "}")
				} else {
					element.Send(msgOthers)
				}
			}
		}
	}(emisor_id, receiver_id, msgOthers)
}

func RandomValue(min, max int) int {
	return rand.Intn(max-min) + min
}
func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}

var (
	RealKeys = make(map[string]*TKeyObject)
	maps     = make([]*TMap, 0, 0)
)

func initDB(db *sql.DB) {
	stmt, e := db.Prepare("UPDATE `users` SET `isAlive`=0;")
	if e != nil {
		panic(e.Error())
	}
	defer stmt.Close()
	stmt.Exec()

	stmt, e = db.Prepare("SELECT o.color, o.id, om.posX, om.posY FROM object_map om INNER JOIN object o ON o.id = om.id_obj WHERE om.id_map=1;")
	if e != nil {
		panic(e.Error())
	}
	defer stmt.Close()
	rows, _ := stmt.Query()

	for rows.Next() {
		var color, id, posX, posY []byte // color, id, posX, posY
		rows.Scan(&color, &id, &posX, &posY)
		fmt.Println(string(color), string(id))
		px, _ := strconv.ParseFloat(string(posX), 64)
		py, _ := strconv.ParseFloat(string(posY), 64)
		i, _ := strconv.Atoi(string(id))
		RealKeys[string(color)] = &TKeyObject{Id: i, PosX: px, PosY: py, Color: string(color)}
	}
	rows.Close()
	fmt.Println("REAL KEYS: ", len(RealKeys))

	// READ MAP
	stmt, e = db.Prepare("SELECT * FROM `map` WHERE `id`= 1;")
	if e != nil {
		panic(e.Error())
	}
	defer stmt.Close()
	rows, _ = stmt.Query()

	for rows.Next() {
		var id, mapFields, width, height []byte // id, mapFields, width, height
		rows.Scan(&id, &mapFields, &width, &height)
		i, _ := strconv.Atoi(string(id))
		w, _ := strconv.Atoi(string(width))
		h, _ := strconv.Atoi(string(height))
		maps = append(maps, &TMap{Id: i, MapFields: string(mapFields), Width: w, Height: h})
	}
	rows.Close()
	fmt.Println("MAPS: ", len(maps))
	fmt.Println(maps[0])
}

func main() {
	rand.Seed(time.Now().Unix())
	db, err := sql.Open("mysql", "root:@/tfg_gis")
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	initDB(db)

	server := New(CONN_HOST + ":" + CONN_PORT)

	server.OnNewClient(func(c *Client) {
		// new client connected
		id, _ := strconv.Atoi(strings.Split(c.conn.RemoteAddr().String(), ":")[1]) // Assign port to Client
		c.port = id
		server.clients[id] = c
		fmt.Println("Nuevo cliente con id =  ", id)
	})
	server.OnNewMessage(func(c *Client, message string) {
		// new message received
		data := map[string]interface{}{}
		dec := json.NewDecoder(strings.NewReader(message))
		_ = dec.Decode(&data)
		jq := jsonq.NewQuery(data)
		action, _ := jq.String("Action")

		action = string(action)

		log.Print("Accion => ", action)
		if "initWName" == action {
			if c.username == "" {
				username, _ := jq.String("Name")
				username = string(username)
				c.username = username
				user, e := db.Prepare("SELECT `posX`, `posY` FROM `users` WHERE `username`='" + username + "'")
				if e != nil {
					panic(e.Error())
				}
				defer user.Close()
				rows, _ := user.Query()
				var posX float64 = 320
				var posY float64 = 320
				exists := false
				for rows.Next() {
					var px, py []byte
					rows.Scan(&px, &py)
					posX, _ = strconv.ParseFloat(string(px), 64)
					posY, _ = strconv.ParseFloat(string(py), 64)
					exists = true
					break
				}
				rows.Close()

				objUser := &TObjectUser{Id: c.port, PosX: posX, PosY: posY}

				var str string
				if exists {
					str = "UPDATE `users` SET `port`=" + strconv.Itoa(c.port) + ", `isAlive`=1 WHERE `username`='" + username + "';"
				} else {
					str = "INSERT INTO `users` (`port`, `username`) VALUES ('" + strconv.Itoa(c.port) + "', '" + username + "');"
				}
				stmt, ee := db.Prepare(str)
				if ee != nil {
					panic(ee.Error())
				}
				defer stmt.Close()
				stmt.Exec()

				users := make(map[string]*TObjectUser)
				users_, ee := db.Prepare("SELECT `port`, `posX`, `posY` FROM `users` WHERE `isAlive`=1 AND `port` NOT IN (" + strconv.Itoa(c.port) + ")")
				if ee != nil {
					panic(ee.Error())
				}
				defer users_.Close()
				rows, _ = users_.Query()
				for rows.Next() {
					var pt, px, py []byte
					rows.Scan(&pt, &px, &py)
					prt_str := string(pt)
					prt_int, _ := strconv.Atoi(prt_str)
					posX, _ := strconv.ParseFloat(string(px), 64)
					posY, _ := strconv.ParseFloat(string(py), 64)
					users[prt_str] = &TObjectUser{Id: prt_int, PosX: posX, PosY: posY}
				}
				rows.Close()

				toSend := SendMap{Action: "sendMap", Id: objUser.Id, Map: maps[objUser.Map], X: objUser.PosX, Y: objUser.PosY, Users: users}
				msg, _ := toSend.MarshalJSON()
				c.Send(string(msg))
				mssg := "{\"Action\":\"new\",\"Id\":" + strconv.Itoa(c.port) + ",\"PosX\":" + FloatToString(objUser.PosX) + ",\"PosY\":" + FloatToString(objUser.PosY) + "}"
				go func(mssg string, port int) {
					for index, element := range server.clients {
						if index != port {
							element.Send(mssg)
						}
					}
				}(mssg, c.port)
			}
		} else if "move" == action {
			go func(port int, msg string) {
				for index, element := range server.clients {
					if index != port {
						element.Send(msg)
					}
				}
			}(c.port, message)
			go func() {
				px, _ := jq.Float("Pos", "X")
				px_ := FloatToString(px)
				py, _ := jq.Float("Pos", "Y")
				py_ := FloatToString(py)
				stmt, e := db.Prepare("UPDATE `users` SET `port`=" + strconv.Itoa(c.port) + ",`posX`=" + px_ + ",`posY`=" + py_ + " WHERE `port`=" + strconv.Itoa(c.port) + ";")
				if e != nil {
					panic(e.Error())
				}
				defer stmt.Close()
				stmt.Exec()
			}()
		} else if "fight" == action {
			sendFight(db, c.port, jq, server)
			return
		} else if "finishBattle" == action {
			emisor_id := c.port
			receiver_id, _ := jq.Int("Id_enemy")
			stmt, e := db.Prepare("SELECT `port`, `rollDice` FROM `users` WHERE `port`=" + strconv.Itoa(emisor_id) + " or `port`=" + strconv.Itoa(receiver_id) + ";")
			if e != nil {
				panic(e.Error())
			}
			defer stmt.Close()
			rows, _ := stmt.Query()
			var port int = 0
			var emisor_roll int = -1
			var receiver_roll int = -1
			for rows.Next() {
				var p, r []byte
				rows.Scan(&p, &r)
				port, _ = strconv.Atoi(string(p))
				if port == emisor_roll {
					emisor_roll, _ = strconv.Atoi(string(r))
				} else if port == receiver_roll {
					receiver_roll, _ = strconv.Atoi(string(r))
				}
			}
			rows.Close()

			winner := -1
			valueC := -1
			valueE := -1
			if receiver_id == -1 {
				winner = emisor_id
				valueE = emisor_roll
			} else if emisor_id == -1 {
				winner = receiver_id
				valueC = receiver_roll
			} else if emisor_roll > receiver_roll {
				winner = emisor_id
				valueE = emisor_roll
				valueC = receiver_roll
			} else if receiver_roll > emisor_roll {
				winner = receiver_id
				valueE = emisor_roll
				valueC = receiver_roll
			}
			c.Send("{\"Action\":\"finishBattle\",\"ValueClient\":" + strconv.Itoa(valueC) + ",\"ValueEnemy\":" + strconv.Itoa(valueE) + ",\"Winner\":" + strconv.Itoa(winner) + "}")
		} else if "getObj" == action {

		} else if "freeObj" == action {

		} else if "exit" == action {
			server.onClientConnectionClosed(c, nil)
			return
		}
	})
	server.OnClientConnectionClosed(func(c *Client, err error) {
		// connection with client lost or desconnected
		fmt.Println("Connection closed")
		id := c.port
		username := c.username
		delete(server.clients, id)
		msg := "{\"Action\":\"exit\",\"Id\":" + strconv.Itoa(id) + "}"
		go func() {
			for _, element := range server.clients {
				element.Send(msg)
			}
		}()
		stmt, eee := db.Prepare("UPDATE `users` SET `isAlive`=0 WHERE `username`='" + username + "';")
		if eee != nil {
			panic(eee.Error())
		}
		defer stmt.Close()
		stmt.Exec()
	})

	server.Listen()
}
