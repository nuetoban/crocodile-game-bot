/*
 * This file is part of Crocodile Game Bot.
 * Copyright (C) 2019  Viktor
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package storage

import (
	"encoding/json"
	"strconv"

	"github.com/gomodule/redigo/redis"
	"github.com/nuetoban/crocodile-game-bot/crocodile"
)

type Redis struct {
	Pool *redis.Pool
}

func (r *Redis) SaveMachineState(m crocodile.Machine) error {
	j, err := json.Marshal(m)
	if err != nil {
		return err
	}

	conn := r.Pool.Get()
	defer conn.Close()

	key := "machine/" + strconv.Itoa(int(m.ChatID))
	conn.Do("SET", key, string(j))
	conn.Do("EXPIRE", key, "86400")

	return nil
}

// LookupForMachine will take machine from Redis and unmarshal json to m argument
func (r *Redis) LookupForMachine(m *crocodile.Machine) error {
	conn := r.Pool.Get()
	defer conn.Close()

	resp, err := conn.Do("GET", "machine/"+strconv.Itoa(int(m.ChatID)))
	if err != nil {
		return err
	}
	if r, ok := resp.([]byte); ok {
		err = json.Unmarshal(r, m)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewRedis(p *redis.Pool) *Redis {
	return &Redis{Pool: p}
}
