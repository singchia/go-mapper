package procedure

/*
* 		@author zhzhai
*
*		data struction in redis
*
* 	m(variable key)     		map(string type value)
*		m1				map1
*		m2				map2
*		m3				map2
*  		m4				map2
*
*	map(variable key) 		ms(set type value)
*   map1 				{m1}
*   map2 				{m2, m3, m4}
*
*	go-mapper:maps(const key)			{map1, map2} (set type value)
*	go-mapper:ms(const key)			{m1, m2, m3, m4} (set type value)
*
*   map:per(partial variable key)	per(string type value)
* 	map1:per 			 	per1
*	map2:per 				per2
*
*	p(variable key)			per(string type value)
*	p1 						per1
*	p2 						per1
*	p3 						per1
*	p4 						per1
*	p5 						per2
*
*  	per(variable key)		ps(set type value)
*	per1 					{p1, p2, p3, p4, p5}
*	per2 					{p5}
*
*	go-mapper:pers(const key)			{per1, per2}(set type value)
*	go-mapper:ps(const key)				{p1, p2, p3, p4, p5}(set type value)
*
*
*		return errno meaning
*
*	return number		meaning
*	0			nothing wrong
*	40010		something fucked up
*	40011		ISMEMBER operation error
*	40012		SADD operation error
*	40021		already exists
*	40022		contradiction exists
*	40023		entity not exists
**/

//CREATE
/*
*	create a {map}
*	if {map} exists, return 40021
*	else put map into "go-mapper:maps"
*	KEYS[1]: pre create {map}
**/
var LuaCreateMap = `
	local rst = redis.call("SISMEMBER", "go-mapper:maps", KEYS[1])
	if rst == 1 then 
		--[[ pre create {map} alreay in "go-mapper:maps" ]]--
		return 40021
	elseif rst == 0 then 
		rst = redis.call("SADD", "go-mapper:maps", KEYS[1])
		if rst == 1 then
			return 0
		elseif rst == 0 then
			--[[ contradiction exists ]]
			return 40022
		else 
			--[[ something fucked up ]]--
			return 40012
		end
	else
		--[[ something fucked up ]]--
		return 40011
	end
`

/*
*	create a {m}
*	if {m} exists, return 40021
*	else put {m} into "go-mapper:ms"
*	KEYS[1]: pre create {m}
**/
var LuaCreateM = `
	local rst = redis.call("SISMEMBER", "go-mapper:ms", KEYS[1])
	if rst == 1 then 
		return 40021
	elseif rst == 0 then 
		rst = redis.call("SADD", "go-mapper:ms", KEYS[1])
		if rst == 1 then
			return 0
		elseif rst == 0 then
			return 40022
		else 
			return 40012
		end
	else
		return 40011
	end
`

/*
*	create a {per}
*	if {per} exists, return 40021
*	else create {per} and put {per} into "go-mapper:pers"
*	KEYS[1]: pre create {per}
**/
var LuaCreatePer = `
	local rst = redis.call("SISMEMBER", "go-mapper:pers", KEYS[1])
	if rst == 1 then 
		return 40021
	elseif rst == 0 then 
		rst = redis.call("SADD", "go-mapper:pers", KEYS[1])
		if rst == 1 then 
			return 0
		elseif rst == 0 then
			return 40022
		else 
			return 40012
		end
	else
		return 40011
	end
`

/*
*	create a {p}
*	if {p} exists, return 40021
*	else put {p} into "go-mapper:ps"
*	KEYS[1]: pre create {p}
**/
var LuaCreateP = `
	local rst = redis.call("SISMEMBER", "go-mapper:ps", KEYS[1])
	if rst == 1 then 
		return 40021
	elseif rst == 0 then 
		rst = redis.call("SADD", "go-mapper:ps", KEYS[1])
		if rst == 1 then 
			return 0
		elseif rst == 0 then
			return 40022
		else 
			return 40012
		end
	else
		return 40011
	end
`

//UPDATE
/*
*	update operations just do assignment assign m to map, assign p to per,
*	assign map to per, if assignment exists, then reassign it.
**/

/*
*	assign {m} to {map}
*	if {m} or {map} not exists, return 40023
*	if assignment exists & {m} !=> {map} then reassign {m} => {map}
*	KEYS[1]: pre assign {m}, ARGV[1]: pre assign {map}
**/
var LuaAssignM2Map = `
	local rst = redis.call("SISMEMBER", "go-mapper:ms", KEYS[1])
	if rst ~= 1 then 
		return 40023
	end

	rst = redis.call("SISMEMBER", "go-mapper:maps", ARGV[1])
	if rst ~= 1 then 
		return 40023
	end

	rst = redis.call("GET", KEYS[1])
	if rst ~= false and rst ~= ARGV[1] then 
		--[[ {m} was assigned to another map ]]--
		rst = redis.call("SREM", rst, KEYS[1])
		if rst ~= 1 then 
			redis.log(redis.LOG_WARNING, "LuaAssignM2Map | SREM m in map error, ret: ", rst, ", type: ", type(rst))
			return 40013
		end
	end

	rst = redis.call("SADD", ARGV[1], KEYS[1])
	if rst == 1 or rst == 0 then 
		--[[ SADD succeed ]]--
		rst = redis.call("SET", KEYS[1], ARGV[1])
		if type(rst) == "table" and rst.ok == "OK" then 
			--[[ SET succeed ]]--
			return 0
		else
			--[[ not occured while testing ]]--
			return 40031
		end
	else 
		--[[ SADD {m} into {map} failed ]]--
		redis.log(redis.LOG_WARNING, "LuaAssginM2Map | SADD m to map error, ret: ", rst, " ,type: ", type(rst))
		return 40012
	end
`

/*
*	assign multi {m} to {map}
*	if {map} not exists, return 40023
*	if assignment exists & {m} !=> {map} then reassign {m} => {map}
*	KEYS[1]: multi {m}s splited with " ", ARGV[1]: pre assign {map}
**/
var LuaAssignMultiM2Map = `
	local rst = redis.call("SISMEMBER", "go-mapper:maps", ARGV[1])
	if rst ~= 1 then 
		return 40023
	end
	
	for word in KEYS[1]:gmatch("([^%s]+)") do 
		rst = redis.call("SISMEMBER", "go-mapper:ms", word)
		if rst == 1 then 
			rst = redis.call("GET", word)
			if rst ~= false and rst ~= ARGV[1] then 
				--[[ {m} was assigned to another map ]]--
				rst = redis.call("SREM", rst, word)
				if rst ~= 1 then 
					redis.log(redis.LOG_WARNING, "LuaAssignMultiM2Map | SREM m in map error, ret: ", rst, ", type: ", type(rst))
					return 40013
				end
			end

			rst = redis.call("SADD", ARGV[1], word)
			if rst == 1 or rst == 0 then 
				rst = redis.call("SET", word, ARGV[1])
				if type(rst) == "table" and rst.ok == "OK" then 
					--[[ continue ]]--
				else
					--[[ not occured while testing ]]--
					return 40031
				end
			else 
				redis.log(redis.LOG_WARNING, "LuaAssginMultiM2Map | SADD m to map error, ret: ", rst, " ,type: ", type(rst))
				return 40012
			end
		end
	end
	return 0
`

/*
*	assign {p} to {per}
*	if {p} or {per} not exists, return 40023
*	if assignment exists & {p} !=> {per} then reassign {p} => {per}
*	KEYS[1]: pre assign {p}, ARGV[1]: pre assign {per}
**/
var LuaAssignP2Per = `
	local rst = redis.call("SISMEMBER", "go-mapper:ps", KEYS[1])
	if rst ~= 1 then 
		return 40023
	end

	rst = redis.call("SISMEMBER", "go-mapper:pers", ARGV[1])
	if rst ~= 1 then 
		return 40023
	end

	rst = redis.call("GET", KEYS[1])
	if rst ~= false and rst ~= ARGV[1] then 
		--[[ {p} was assigned to another per ]]--
		rst = redis.call("SREM", rst, KEYS[1])
		if rst ~= 1 then 
			redis.log(redis.LOG_WARNING, "LuaAssignP2Per | SREM p in per error, ret: ", rst, ", type: ", type(rst))
			return 40013
		end
	end

	rst = redis.call("SADD", ARGV[1], KEYS[1])
	if rst == 1 or rst == 0 then 
		--[[ SADD succeed ]]--
		rst = redis.call("SET", KEYS[1], ARGV[1])
		if type(rst) == "table" and rst.ok == "OK" then 
			--[[ SET succeed ]]--
			return 0
		else
			--[[ not occured while testing ]]--
			return 40031
		end
	else 
		redis.log(redis.LOG_WARNING, "LuaAssginP2Per | SADD p to per error, ret: ", rst, ", type: ", type(rst)) return 40012
	end
`

/*
*	assign multi {p}s to {per}
*	if {per} not exists, return 40023
*	if assignment exists & {p} !=> {per} then reassign {p} => {per}
*	KEYS[1]: multi ps splited with " ", ARGV[1]: pre assign {per}
**/
var LuaAssignMultiP2Per = `
	local rst = redis.call("SISMEMBER", "go-mapper:pers", ARGV[1])
	if rst ~= 1 then 
		return 40023
	end
	
	for word in KEYS[1]:gmatch("([^%s]+)") do 
		rst = redis.call("SISMEMBER", "go-mapper:ps", word)
		if rst == 1 then 
			rst = redis.call("GET", word)
			if rst ~= false and rst ~= ARGV[1] then 
				rst = redis.call("SREM", rst, word)
				if rst ~= 1 then 
					redis.log(redis.LOG_WARNING, "LuaAssignMultiP2Per | SREM p in per error, ret: ", rst, ", type: ", type(rst))
					return 40013
				end
			end

			rst = redis.call("SADD", ARGV[1], word)
			if rst == 1 or rst == 0 then 
				rst = redis.call("SET", word, ARGV[1])
				if type(rst) == "table" and rst.ok == "OK" then 
					--[[ continue ]]--
				else
					--[[ not occured while testing ]]--
					return 40031
				end
			else 
				redis.log(redis.LOG_WARNING, "LuaAssginMultiP2Per | SADD p to per error, ret: ", rst, " ,type: ", type(rst))
				return 40012
			end
		end
	end
	return 0
`

/*
*	assign {map} to {per}
*	if {map} or {per} not exists, return 40023
*	if assignment exists & {map} !=> {per} then reassign {map} => {per}
*	KEYS[1]: pre assign {map}, ARGV[1]: pre assign {per}
**/
var LuaAssignMap2Per = `
	local rst = redis.call("SISMEMBER", "go-mapper:maps", KEYS[1])
	if rst ~= 1 then
		return 40023
	end

	rst = redis.call("SISMEMBER", "go-mapper:pers", ARGV[1])
	if rst ~= 1 then 
		return 40023
	end

	rst = redis.call("SET", KEYS[1]..":per", ARGV[1])
	if type(rst) == "table" and rst.ok == "OK" then 
		return 0
	else 
		redis.log(redis.LOG_WARNING, "LuaAssignMap2Per | SET map to per error, ret: ", rst, ", type: ", type(rst))
		return 40031
	end
`

/*
*	move random n {p}s from per1 to per2
*	if per1 or per2 not exists return 40023
*	if n > len(per1) then move all {p}s to per2
*	KEYS[1]: per1, ARGV[1]: n, ARGV[2]: per2
*	return: succeed number
**/
var LuaMovePs = `
	local rst = redis.call("SISMEMBER", "go-mapper:pers", KEYS[1])
	if rst ~= 1 then 
		return 40023
	end

	rst = redis.call("SISMEMBER", "go-mapper:pers", ARGV[2])
	if rst ~= 1 then 
		return 40023
	end

	local num = tonumber(ARGV[1])	
	rst = redis.call("SMEMBERS", KEYS[1])
	if num > #rst then 
		num = #rst
	end

	local index = 0
	for k, v in pairs(rst) do 
		local ret = redis.call("SREM", KEYS[1], v)
		if ret == 1 then 
			ret = redis.call("SADD", ARGV[2], v)
			if ret == 1 then 
				ret = redis.call("SET", v, ARGV[2])
				if type(ret) == "table" and ret.ok == "OK" then 
					index = index + 1
				end
			end

		end
		if index == num then 
			break
		end
	end
	return index
`

/*
*	move random n {m}s from map1 to map2
*	if map1 or map2 not exists return 40023
*	if n > len(map1) then move all {p}s to map2
*	KEYS[1]: map1, ARGV[1]: n, ARGV[2]: map2
*	return: succeed number
**/
var LuaMoveMs = `
	local rst = redis.call("SISMEMBER", "go-mapper:maps", KEYS[1])
	if rst ~= 1 then 
		return 40023
	end

	rst = redis.call("SISMEMBER", "go-mapper:maps", ARGV[2])
	if rst ~= 1 then 
		return 40023
	end

	local num = tonumber(ARGV[1])	
	rst = redis.call("SMEMBERS", KEYS[1])
	if num > #rst then 
		num = #rst
	end

	local index = 0
	for k, v in pairs(rst) do 
		local ret = redis.call("SREM", KEYS[1], v)
		if ret == 1 then 
			ret = redis.call("SADD", ARGV[2], v)
			if ret == 1 then 
				ret = redis.call("SET", v, ARGV[2])
				if type(ret) == "table" and ret.ok == "OK" then 
					index = index + 1
				end
			end
		end
		if index == num then 
			break
		end
	end
	return index
`

/*
*	move fixed n {p}s from per1 to per2
*	if per1 or per2 not exists return 40023
*	move per1-existing {p}s to per2
*	KEYS[1]: per1, ARGV[1]: {p}s, ARGV[2]: per2
*	return succeed number
**/
var LuaMoveFixedPs = `
	local rst = redis.call("SISMEMBER", "go-mapper:pers", KEYS[1])
	if rst ~= 1 then 
		return 40023
	end

	rst = redis.call("SISMEMBER", "go-mapper:pers", ARGV[2])
	if rst ~= 1 then 
		return 40023
	end

	local words = {}
	for word in ARGV[1]:gmatch("([^%s]+)") do 
		rst = redis.call("SISMEMBER", KEYS[1], word)
		if rst == 1 then 
			table.insert(words, word) 
		end
	end

	local index = 0
	for k, v in pairs(words) do 
		rst = redis.call("SREM", KEYS[1], v)
		if rst == 1 then 
			rst = redis.call("SADD", ARGV[2], v)
			if rst == 1 then 
				rst = redis.call("SET", v, ARGV[2])
				if type(rst) == "table" and rst.ok == "OK" then 
					index = index + 1
				end
			end
		end
	end
	return index
`

/*
*	move fixed {m}s from map1 to map2
*	if map1 or map2 not exists return 40023
*	move map1-existing {m}s to map2
*	KEYS[1]: map1, ARGV[1]: {m}s splited with " ", ARGV[2]: map2
*	return succeed number
**/
var LuaMoveFixedMs = `
	local rst = redis.call("SISMEMBER", "go-mapper:maps", KEYS[1])
	if rst ~= 1 then 
		return 40023
	end

	rst = redis.call("SISMEMBER", "go-mapper:maps", ARGV[2])
	if rst ~= 1 then 
		return 40023
	end

	local words = {}
	for word in ARGV[1]:gmatch("([^%s]+)") do 
		rst = redis.call("SISMEMBER", KEYS[1], word)
		if rst == 1 then 
			table.insert(words, word) 
		end
	end

	local index = 0
	for k, v in pairs(words) do 
		rst = redis.call("SREM", KEYS[1], v)
		if rst == 1 then 
			rst = redis.call("SADD", ARGV[2], v)
			if rst == 1 then 
				rst = redis.call("SET", v, ARGV[2])
				if type(rst) == "table" and rst.ok == "OK" then 
					index = index + 1
				end
			end
		end
	end
	return index
`

/*
* 	book n {p}s from {per}
*	if {per}:booked exists return 40024
*	if n > len(per) then book all {p}s in {per}
*	KEYS[1]: {per}, ARGV[1]: n
*	return succeed number
**/
var LuaBookPs = `
	local rst = redis.call("EXISTS", KEYS[1]..":booked")
	if rst == 1 then 
		return 40024	
	end

	rst = redis.call("SISMEMBER", "go-mapper:pers", KEYS[1])
	if rst ~= 1 then 
		return 40023
	end

	local num = tonumber(ARGV[1])
	rst = redis.call("SMEMBERS", KEYS[1])
	if #rst == 0 then 
		return 40025
	end
	if num > #rst then 
		num = #rst
	end

	local index = 0
	for k, v in pairs(rst) do 
		local ret = redis.call("SREM", KEYS[1], v)
		if ret == 1 then 
			ret = redis.call("SADD", KEYS[1]..":booked", v)
			if ret == 1 then 
				ret = redis.call("SET", v, KEYS[1]..":booked")
				if type(ret) == "table" and ret.ok == "OK" then 
					index = index + 1
				end
			end
		end
		if index == num then 
			break
		end
	end

	if index > 0 then 
		rst = redis.call("SADD", "go-mapper:pers", KEYS[1]..":booked")
		if rst ~= 1 then 
			return 40012
		end
	end
	return index
`

/*
* 	book fixed {p}s from {per}
*	if {per}:booked exists return 40024
*	KEYS[1]: {per}, ARGV[1]: multi {p}s splited with " "
*	return succeed number
**/
var LuaBookFixedPs = `
	local rst = redis.call("EXISTS", KEYS[1]..":booked")
	if rst == 1 then 
		return 40024	
	end

	rst = redis.call("SISMEMBER", "go-mapper:pers", KEYS[1])
	if rst ~= 1 then 
		return 40023
	end

	local words = {}
	for word in ARGV[1]:gmatch("([^%s]+)") do 
		rst = redis.call("SISMEMBER", KEYS[1], word)
		if rst == 1 then 
			table.insert(words, word) 
		end
	end

	local index = 0
	for k, v in pairs(words) do 
		local ret = redis.call("SREM", KEYS[1], v)
		if ret == 1 then 
			ret = redis.call("SADD", KEYS[1]..":booked", v)
			if ret == 1 then 
				ret = redis.call("SET", v, KEYS[1]..":booked")
				if type(ret) == "table" and ret.ok == "OK" then 
					index = index + 1
				end
			end
		end
	end
	if index > 0 then 
		rst = redis.call("SADD", "go-mapper:pers", KEYS[1]..":booked")
		if rst ~= 1 then 
			return 40012
		end
	end
	return index
`

/*
* 	take {p}s from {per1}:booked to {per2}
*	if {per1}:booked not exists return 40023
*	KEYS[1]: {per1}, ARGV[1]: {per2}
*	return succeed number
**/
var LuaTakePs = `
	local rst = redis.call("EXISTS", KEYS[1]..":booked")
	if rst ~= 1 then 
		return 40023	
	end

	rst = redis.call("SISMEMBER", "go-mapper:pers", KEYS[1])
	if rst ~= 1 then 
		return 40023
	end

	rst = redis.call("SISMEMBER", "go-mapper:pers", ARGV[1])
	if rst ~= 1 then 
		return 40023
	end

	rst = redis.call("SMEMBERS", KEYS[1]..":booked")

	local index = 0
	for k, v in pairs(rst) do 
		local ret = redis.call("SREM", KEYS[1]..":booked", v)
		if ret == 1 then 
			ret = redis.call("SADD", ARGV[1], v)
			if ret == 1 then 
				ret = redis.call("SET", v, ARGV[1]) 
				if type(ret) == "table" and ret.ok == "OK" then 
					index = index + 1
				end
			end
		end
	end
	if index > 0 then 
	rst = redis.call("SREM", "go-mapper:pers", KEYS[1]..":booked")
		if rst ~= 1 then 
			return 40012
		end
	end
	return index
`

/*
* 	book n {m}s from {map}
*	if {map}:booked exists return 40024
*	if n > len(map) then book all {m}s in {map}
*	KEYS[1]: {map}, ARGV[1]: n
*	return succeed number
**/
var LuaBookMs = `
	local rst = redis.call("EXISTS", KEYS[1]..":booked")
	if rst == 1 then 
		return 40024	
	end

	rst = redis.call("SISMEMBER", "go-mapper:maps", KEYS[1])
	if rst ~= 1 then 
		return 40023
	end

	local num = tonumber(ARGV[1])
	rst = redis.call("SMEMBERS", KEYS[1])
	if #rst == 0 then 
		return 40025
	end

	if num > #rst then 
		num = #rst
	end

	local index = 0
	for k, v in pairs(rst) do 
		local ret = redis.call("SREM", KEYS[1], v)
		if ret == 1 then 
			ret = redis.call("SADD", KEYS[1]..":booked", v)
			if ret == 1 then 
				ret = redis.call("SET", v, KEYS[1]..":booked")
				if type(ret) == "table" and ret.ok == "OK" then 
					index = index + 1
				end
			end
		end
		if index == num then 
			break
		end
	end
	if index > 0 then 
		rst = redis.call("SADD", "go-mapper:maps", KEYS[1]..":booked")
		if rst ~= 1 then 
			return 40012
		end
	end
	return index
`

/*
* 	book fixed {m}s from {map}
*	if {map}:booked exists return 40024
*	KEYS[1]: {map}, ARGV[1]: multi {m}s splited with " "
*	return succeed number
**/
var LuaBookFixedMs = `
	local rst = redis.call("EXISTS", KEYS[1]..":booked")
	if rst == 1 then 
		return 40024	
	end

	rst = redis.call("SISMEMBER", "go-mapper:maps", KEYS[1])
	if rst ~= 1 then 
		return 40023
	end

	local words = {}
	for word in ARGV[1]:gmatch("([^%s]+)") do 
		rst = redis.call("SISMEMBER", KEYS[1], word)
		if rst == 1 then 
			table.insert(words, word) 
		end
	end

	local index = 0
	for k, v in pairs(words) do 
		local ret = redis.call("SREM", KEYS[1], v)
		if ret == 1 then 
			ret = redis.call("SADD", KEYS[1]..":booked", v)
			if ret == 1 then 
				ret = redis.call("SET", v, KEYS[1]..":booked")
				if type(ret) == "table" and ret.ok == "OK" then 
					index = index + 1
				end
			end
		end
	end
	if index > 0 then 
		rst = redis.call("SADD", "go-mapper:maps", KEYS[1]..":booked")
		if rst ~= 1 then 
			return 40012
		end
	end
	return index
`

/*
* 	take {m}s from {map1}:booked to {map2}
*	if {map1}:booked not exists return 40023
*	KEYS[1]: {map1}, ARGV[1]: {map2}
*	return succeed number
**/
var LuaTakeMs = `
	local rst = redis.call("EXISTS", KEYS[1]..":booked")
	if rst ~= 1 then 
		return 40023	
	end

	rst = redis.call("SISMEMBER", "go-mapper:maps", KEYS[1])
	if rst ~= 1 then 
		return 40023
	end

	rst = redis.call("SISMEMBER", "go-mapper:maps", ARGV[1])
	if rst ~= 1 then 
		return 40023
	end

	rst = redis.call("SMEMBERS", KEYS[1]..":booked")

	local index = 0
	for k, v in pairs(rst) do 
		local ret = redis.call("SREM", KEYS[1]..":booked", v)
		if ret == 1 then 
			ret = redis.call("SADD", ARGV[1], v)
			if ret == 1 then 
				ret = redis.call("SET", v, ARGV[1]) 
				if type(ret) == "table" and ret.ok == "OK" then 
					index = index + 1
				end
			end
		end
	end
	if index > 0 then 
		rst = redis.call("SREM", "go-mapper:maps", KEYS[1]..":booked")
		if rst ~= 1 then 
			return 40012
		end
	end
	return index
`

//DELETE
/*
*	delete operation to delete enties or assignment
*	delete enties will delete attached assignment
*	delete assignment does't effect others
**/

/*
*	delete p will: SREM go-mapper:ps p, DEL p, SREM assigned-per p
*	KEYS[1]: {p}
**/
var LuaDeleteP = `
	local rst = redis.call("GET", KEYS[1])
	local ret = 0
	if rst == false or type(rst) ~= "string" then 
		ret = 40032
	else 
		rst = redis.call("SREM", rst, KEYS[1])	
		if rst ~= 1 and rst ~= 0 then 
			ret = 40023
		end
	end

	rst = redis.call("DEL", KEYS[1])
	if rst ~= 1 and rst ~= 1 then 
		ret = 40033 
	end

	rst = redis.call("SREM", "go-mapper:ps", KEYS[1])
	if rst ~= 1 and rst ~= 0 and rst ~= 0 then 
		return ret
	elseif rst ~= 1 and rst ~= 0 then 
		return 40023
	else
		return ret 
	end
`

/*
*	delete {p}s will: SREM go-mapper:ps {p} in {p}s, DEL {p} in {p}s, SREM assigned-per {p} in {p}s
*	KEYS[1]: multi {p}s splited with " "
**/
var LuaDeletePs = `
	local ret = 0
	for word in KEYS[1]:gmatch("([^%s]+)") do 
		local rst = redis.call("GET", word)
		if rst == false or type(rst) ~= "string" then 
			ret = 40032
		else 
			rst = redis.call("SREM", rst, word)	
			if rst ~= 1 and rst ~= 0 then 
				ret = 40023
			end
		end

		rst = redis.call("DEL", word)
		if rst ~= 1 and rst ~= 0 then 
			ret = 40033 
		end

		rst = redis.call("SREM", "go-mapper:ps", word)
		if rst ~= 1 and rst == 0 then 
			ret = 40023
		end
	end
	return ret
`

/*
*	delete per will: DEL p in per, SREM p in ps, DEL {per}, SREM {per} in pers
*	KEYS[1]: {per}
*	NOTE: this operation is quiet dangerous, it may cause a {map} assigned-less
**/
var LuaDeletePer = `
	local rst = redis.call("SMEMBERS", KEYS[1])
	local ret = 0
	for k, v in pairs(rst) do
		local rstDel = redis.call("DEL", v)
		if rstDel ~= 1 and rstDel ~= 0  then 
			ret = 40033
		end
		local rstSrem = redis.call("SREM", "go-mapper:ps", v)
		if rstSrem ~= 1 and rstSrem ~= 0 then 
			ret = 40023
		end
	end

	rst = redis.call("DEL", KEYS[1])
	if rst ~= 1 and rst ~= 0  then 
		ret = 40033
	end

	rst = redis.call("SREM", "go-mapper:pers", KEYS[1])
	if rst ~= 1 and rst ~= 0 then 
		ret = 40023
	end
	return ret
`

/*
*	delete {m} will: SREM go-mapper:ms {m}, DEL {m}, SREM assigned-map {m}
*	KEYS[1]: {m}
**/
var LuaDeleteM = `
	local rst = redis.call("GET", KEYS[1])
	local ret = 0
	if rst == false or type(rst) ~= "string" then 
		ret = 40032
	else 
		rst = redis.call("SREM", rst, KEYS[1])	
		if rst ~= 1 and rst ~= 0 then 
			ret = 40023
		end
	end

	rst = redis.call("DEL", KEYS[1])
		if rst ~= 1 and rst ~= 0 then 
			ret = 40033 
	end

	rst = redis.call("SREM", "go-mapper:ms", KEYS[1])
	if rst ~= 1 and rst ~= 0 and ret ~= 0 then 
		return ret
	elseif rst ~= 1 and rst ~= 0 then 
		return 40023
	else
		return ret 
	end
`

/*
*	delete {m}s will: SREM go-mapper:ms {m} in {m}s, DEL {m} in {m}s, SREM assigned-map {m} in {m}s
*	KEYS[1]: {m}s
**/
var LuaDeleteMs = `
	local ret = 0
	for word in KEYS[1]:gmatch("([^%s]+)") do 
		local rst = redis.call("GET", word)
		if rst == false or type(rst) ~= "string" then 
			ret = 40032
		else 
			rst = redis.call("SREM", rst, word)	
			if rst ~= 1 and rst ~= 0 then 
				ret = 40023
			end
		end

		rst = redis.call("DEL", word)
			if rst ~= 1 and rst ~= 0 then 
				ret = 40033 
		end

		rst = redis.call("SREM", "go-mapper:ms", word)
		if rst ~= 1 and rst ~= 0 and ret == 0 then 
			ret = 40023
		end
	end
	return ret
`

/*
*	delete {map} will: DEL {m} in {map}, SREM {m} in ms, DEL {map}, SREM {map} in maps, DEL {map}:per
*	KEYS[1]: {map}
**/
var LuaDeleteMap = `
	local rst = redis.call("SMEMBERS", KEYS[1])
	local ret = 0
	for k, v in pairs(rst) do
		local rstDel = redis.call("DEL", v)
		if rstDel ~= 1 and rst ~= 0 then 
			ret = 40033
		end
		local rstSrem = redis.call("SREM", "go-mapper:ms", v)
		if rstSrem ~= 1 and rstSrem ~= 0 then 
			ret = 40023
		end
	end

	rst = redis.call("DEL", KEYS[1])
	if rst ~= 1 and rst ~= 0 then 
		ret = 40033
	end

	rst = redis.call("SREM", "go-mapper:maps", KEYS[1])
	if rst ~= 1 and rst ~= 0 then 
		ret = 40023
	end

	rst = redis.call("DEL", KEYS[1]..":per")
	if rst ~= 1 and rst ~= 0 then 
		ret = 40023
	end
	return ret
`

/*
*	delete {p}'s assignment will: DEL {p}, SREM {p} in {per}
*	KEYS[1]: {p}
**/
var LuaDeletePAssignment = `
	local rst = redis.call("GET", KEYS[1])
	local ret = 0
	if rst == false or type(rst) ~= "string" then 
		ret = 40032
	else 
		local rstDel = redis.call("DEL", KEYS[1])
		if rstDel ~= 1 and rstDel ~= 0 then 
			ret = 40033
		end

		local rstSrem = redis.call("SREM", rst, KEYS[1])
		if rstSrem ~= 1 and rstSrem ~= 0 then 
			ret = 40023
		end
	end
	return ret
`

/*
*	delete {p}s's assignment will: DEL {p}s, SREM {p}s in {per}
*	KEYS[1]: multi {p}s splited with " "
**/
var LuaDeleteMultiPAssignment = `
	local ret = 0
	for word in KEYS[1]:gmatch("([^%s]+)") do 
		local rst = redis.call("GET", word)
		if rst == false or type(rst) ~= "string" then 
			ret = 40032
		else 
			local rstDel = redis.call("DEL", word)
			if rstDel ~= 1 and rstDel ~= 0 then 
				ret = 40033
			end

			rst = redis.call("SREM", rst, word)
			if rst ~= 1 and rst ~= 0 then 
				ret = 40023
			end
		end
	end
	return ret
`

/*
*	delete {m}'s assignment will: DEL {m}, SREM {m} in {map}
*	KEYS[1]: {m}
**/
var LuaDeleteMAssignment = `
	local rst = redis.call("GET", KEYS[1])
	local ret = 0
	if rst == false or type(rst) ~= "string" then 
		ret = 40032
	else 
		local rstDel = redis.call("DEL", KEYS[1])
		if rstDel ~= 1 and rstDel ~= 0 then 
			ret = 40033
		end

		rst = redis.call("SREM", rst, KEYS[1])
		if rst ~= 1 and rst ~= 0 then 
			ret = 40023
		end
	end
	return ret
`

/*
*	delete {m}s's assignment will: DEL {m}s, SREM {m}s in {map}
*	KEYS[1]: multi {m}s splited with " "
**/
var LuaDeleteMultiMAssignment = `
	local ret = 0
	for word in KEYS[1]:gmatch("([^%s]+)") do 
		local rst = redis.call("GET", word)
		if rst == false or type(rst) ~= "string" then 
			ret = 40032
		else 
			local rstDel = redis.call("DEL", word)
			if rstDel ~= 1 and rstDel ~= 0 then 
				ret = 40033
			end

			rst = redis.call("SREM", rst, word)
			if rst ~= 1 and rst ~= 0 then 
				ret = 40023
			end
		end
	end
	return ret
`

/*
*	delete {map}'s assignment will: DEL {map}:per
*	KEYS[1]: {map}
**/
var LuaDeleteMapAssignment = `
	local rst = redis.call("DEL", KEYS[1]..":per")
	if rst ~= 1 and rst ~= 0 then 
		return 40033
	end
	return 0
`

//RETRIEVE
/*
*	retrieve opertations to get data in existing assigment and entities
*	retrieve all {per}s by go-mapper:pers, retrieve all {map}s by go-mapper:maps
*	retrieve {p}s by {per}, retrieve {m}s by {map}
*	retrieve all {p} by go-mapper:ps, retrieve all {m} by go-mapper:ms
*	retrieve {map} by {m}, retrieve {per} by {map}
**/

/*
*	retrieve all {per}s by go-mapper:pers
*	NO KEYS & ARGVS
**/
var LuaRetrieveAllPers = `
	return redis.call("SMEMBERS", "go-mapper:pers")
`

/*
*	retrieve all {map}s by go-mapper:maps
*	NO KEYS & ARGVS
**/
var LuaRetrieveAllMaps = `
	return redis.call("SMEMBERS", "go-mapper:maps")
`

/*
*	retrieve {p} by {per}
*	KEYS[1]: {per}
**/
var LuaRetrievePs = `
	return redis.call("SMEMBERS", KEYS[1])
`

/*
*	retrieve {m} by {map}
*	KEYS[1]: {map}
**/
var LuaRetrieveMs = `
	return redis.call("SMEMBERS", KEYS[1])
`

/*
*	retrieve all {p}s by go-mapper:ps
*	NO KEYS & ARGVS
**/
var LuaRetrieveAllMs = `
	return redis.call("SMEMBERS", "go-mapper:ms")
`

/*
*	retrieve all {m}s by go-mapper:ms
*	NO KEYS & ARGVS
**/
var LuaRetrieveAllPs = `
	return redis.call("SMEMBERS", "go-mapper:ps")
`

/*
*	retrieve {map} by {m}
*	KEYS[1]: {m}
**/
var LuaRetrieveMap = `
	local rst = redis.call("GET", KEYS[1])
	if rst == false or type(rst) ~= "string" then 
		return 40026
	end
	return rst
`

/*
*	retrieve {per} by {p}
*	KEYS[1]: {map}
**/
var LuaRetrievePer = `
	local rst = redis.call("GET", KEYS[1])
	if rst == false or type(rst) ~= "string" then 
		return 40026
	end
	return rst
`

/*
*	retrieve {per} by {map}
*	KEYS[1]: {map}
**/
var LuaRetrievePerByMap = `
	local rst = redis.call("GET", KEYS[1]..":per")
	if rst == false or type(rst) ~= "string" then 
		return 40026
	end
	return rst
`

/*
*	retrieve topology
*	NO KEYS and ARGVS
**/
var LuaRetrieveTopo = `
	local topo = {}
	local maps = {}
	local pers = {}

	local rst = redis.call("SMEMBERS", "go-mapper:maps")
	for k, v in pairs(rst) do 
		local map = {}
		map["map"] = v

		local per = redis.call("GET", v..":per")	
		if type(per) == "string"  then 
			map["c-per"] = per 
		end

		local ms = redis.call("SMEMBERS", v)
		if #rst > 0 then 
			map["c-ms"] = ms
		end
		table.insert(maps, map)
	end
	topo["maps"] = maps

	rst = redis.call("SMEMBERS", "go-mapper:pers")
	for k, v in pairs(rst) do 
		local per = {}
		per["per"] = v
		local ps = redis.call("SMEMBERS", v)
		if #rst > 0 then 
			per["c-ps"] = ps 
		end
		table.insert(pers, per)
	end
	topo["pers"] = pers

	local ms = {}
	rst = redis.call("SMEMBERS", "go-mapper:ms")
	for k, v in pairs(rst) do 
		local m = {}
		local map = redis.call("GET", v)
		if type(map) == "string" then 
			m[v] = map
		else 
			m[v] = "unassigned" 
		end
		table.insert(ms, m)
	end
	if #ms > 0 then 
		topo["ms"] = ms 
	end

	local ps = {}
	rst = redis.call("SMEMBERS", "go-mapper:ps")
	for k, v in pairs(rst) do 
		local p = {}
		local per = redis.call("GET", v)
		if type(per) == "string" then 
			p[v] = per 
		else 
			p[v] = "unassigned" 
		end
		table.insert(ps, p)
	end
	if #ps > 0 then 
		topo["ps"] = ps 
	end

	return cjson.encode(topo)
`

/*
*	retrieve a random {p} by {m}, {m}=>{map}=>{per}=>{p}, the {p} will be took in random
*	KEYS[1]: {m}, ARGV[1]: count of {p} and should be verified by application
**/
var LuaRetrievePsByM = `
	local rst = redis.call("GET", KEYS[1])
	if rst == false or type(rst) ~= "string" then 
		return 40026
	end
	
	rst = redis.call("GET", rst..":per")
	if rst == false or type(rst) ~= "string" then 
		return 40026
	end

	local count = tonumber(ARGV[1])
	rst = redis.call("SRANDMEMBER", rst, count)	
	return rst
`
