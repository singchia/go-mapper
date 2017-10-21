# go-mapper
**go-mapper** maintains mappings between two kinds of groups in redis, an implementation of keeping relational model of data in a NoSQL. **go-mapper** simply use two data types of redis value which are Strings and Sets to hold the model:  


| entity| relation | data type of redis value| entity|
|:---:|:---:|:---:|:---:|
|m|points to|Strings|map|
|map|contains|Sets|m|
|map|points to|Strings|per|
|per|contains|Sets|p|
|p|points to|Strings|per|  

**go-mapper** keeps 4 types of entities which respectively are **m**, **map**, **p** and **per**, and 2 types of relations which respectively are **points to** and **contains**. **map** is the group of **m** and **per** is the group of **p**, if **map** points to **per**, then all **m**s contained in **map** point to **p**s contained in **per**.   
## A use case
Let's see a case in distributed system, assuming you got 4 types of requests like **A**, **B**, **C**, **D**, and 4 nodes **a**, **b**, **c**, **d** to handle those requests, for load balancing, you may want sigle node to handle one specific type of requests, then the relations may look like: 

```
A => a; B => b; C => c; D => d;
```
	
quite good when everything go well, but now node **a** is down, you have to reassign **A** to another node, then it goes to:   

```
A => b; B => b; C => c; D => d; 
```  
now node **b** takes 2 types of requests. troubles never come single, we find that type **C** requests income too fast, but type **D** requests are just on the contrary, we hope node **d** can handle both **C** and **D**:   

```
A => b; B => b; C => c,d; D => d;
```  
what if node **d** went wrong too? we need to iterate all relation to remove the assignment, that's a bad idea at runtime. what about building relations that node points to type of request too?    

```
A => b; B => b; C => c,d; D => d;
b => A,B; c => C; d => C,D;
```  
don't be so happy, now I want all types of requests to share all nodes, you know what, I need to change 8 assignments to achieve that:  
 
```
A => b,c,d; B => b,c,d; C => b,c,d; D => b,c,d; 
b => A,B,C,D; c => A,B,C,D; d => A,B,C,D;
```  
totally a mess! we need to change our mind, **details should depend upon abstractions**, we need to construct a another structure to loose coupling between types of requests and nodes. that's how **go-mapper** came up:    

```
{A, B, C, D} => {b, c, d}
```  
if node **b** is down, just adjust to:   

```
{A, B, C, D} => {c, d}
```  
**solved!**

## Why redis
* main reason, it's stable and fast.
* it's hard to keep consistency when status changed, especially when requests come really fast.
* if not using a back-end component, it wound be triky to implement a distrubuted system with CAP consideration. the atomic status changing requires much more codes and it may be not efficient.
* with redis **go-mapper** can scale out as wish, because all status were keeped in redis, **go-mapper** just proxy the status-changing requests to it.

## Installation
If you don't have the Go development environment installed, visit the [Getting Started](https://golang.org/doc/install) document and follow the instructions. Once you're ready, execute the following command:  

```
go get -u github.com/singchia/go-mapper
```  
then    

```
go install github.com/singchia/go-mapper
```  
you will get a runable binary file named **go-mapper** in **GOBIN**

## How to use
**go-mapper** supports rest api to maintain the mappings, those api manipulate 4 types of entities(**m**, **map**, **p**, **per**) in restful to supply **CURD**(**create**, **update**, **retrieve**, **delete**) operations.    
### CREATE
-----
To create a unassigned m named _**m1**_(not point to any map yet).  [verify with topology](#retrieve)
  
```
> curl -i -X POST http://localhost:1202/ms/m1
```

To create a empty map named _**map1**_(include no m). [verify with topology](#retrieve)

```
> curl -i -X POST http://localhost:1202/maps/map1
```
To create a unassigned p named _**p1**_(not point to any per yet). [verify with topology](#retrieve)

```
> curl -i -X POST http://localhost:1202/ps/p1
```
To create a empty per named _**per1**_(include no p). [verify with topology](#retrieve)

```
> curl -i -X POST http://localhost:1202/pers/per1
```
### UPDATE
-----
To assign _**m1**_ to _**map1**_ and _**map1**_ will contain _**m1**_. [verify with topology](#retrieve)

```
> curl -i -X PUT http://localhost:1202/ms/m1/maps/map1
```

To assign _**p1**_ to _**per1**_ and _**per1**_ will include _**p1**_. [verify with topology](#retrieve)

```
> curl -i -X PUT http://localhost:1202/ps/p1/pers/per1
```
To assign _**map1**_ to _**per1**_. [verify with topology](#retrieve)

```
> curl -i -X PUT http://localhost:1202/maps/map1/pers/per1
```
To move 2 elements from _**map1**_ to _**map2**_(assuming _**map1**_ and _**map2**_ were created), if 2 > count of elements that _**map1**_ contains, then move all. [verify with topology](#retrieve)

```
> curl -i -X PUT http://localhost:1202/maps/map1/maps/map2?count=2
{"count":1}
```
To move 2 elements from _**per1**_ to _**per2**_(assuming _**per1**_ and _**per2**_ were created), if 2 > count of elements that _**per1**_ contains, then move all. [verify with topology](#retrieve)

```
> curl -i -X PUT http://localhost:1202/pers/per1/pers/per2?count=2
{"count":1}
```
To move some specific elements in **http Query** field which should be encoded by **base64** from _**map2**_ to _**map1**_. [verify with topology](#retrieve)

```
> base64 <<< "m1"
bTEK
> curl -i -X PUT http://localhost:1202/maps/map2/maps/map1?elements=bTEK
{"count":1}
```

To move some specific elements in **http Query** field which should be encoded by **base64** from _**per2**_ to _**per1**_. [verify with topology](#retrieve)

```
> base64 <<< "p1"
cDEK
> curl -i -X PUT http://localhost:1202/pers/per2/pers/per1?elements=cDEK
{"count":1}
```
To book 2 elements from _**map1**_, once booked, can't book again before the elements be took. if 2 > count of elements that _**map1**_ contains, then book all. [verify with topology](#retrieve)

```
> curl -i -X PUT http://localhost:1202/maps/map1?count=2
{"count":1}
```
To book specific elements from _**map1**_, once booked, can't book again before the elements be took. [verify with topology](#retrieve)

```
> base64 <<< "m1"
bTEK
> curl -i -X PUT http://localhost:1202/maps/map1?elements=bTEK
{"count":1}
```
To take the elements to _**map2**_ booked from _**map1**_. [verify with topology](#retrieve)

```
> curl -i -X PUT http://localhost:1202/maps/map1/maps/map2
{"count":1}
```
To book 2 elements from _**per1**_, once booked, can't book again before the elements be took. if 2 > count of elements that _**per1**_ contains, then book all. [verify with topology](#retrieve)

```
> curl -i -X PUT http://localhost:1202/pers/per1?count=2
{"count":1}
```
To book specific elements from _**per1**_, once booked, can't book again before the elements be took. [verify with topology](#retrieve)

```
> base64 <<< "p1"
cDEK
> curl -i -X PUT http://localhost:1202/pers/per1?elements=cDEK
{"count":1}
```
To take the elements to _**per2**_ booked from _**per1**_. [verify with topology](#retrieve)

```
> curl -i -X PUT http://localhost:1202/pers/per1/pers/per2
{"count":1}
```
**Note:** operation **book** can be used to lock the resource.

### RETRIEVE
------
**To retrieve go-mapper's topology, c- means correspoding**. [verify with topology](#retrieve)

```
> curl -i -X GET http://localhost:1202/topology
{
  "ms":[{"m1":"map1"}],
  "ps":[{"p1":"per1"}],
  "pers":[
    {"per":"per1","c-ps":["p1"]},
    {"per":"per2","c-ps":{}}],
  "maps":[
    {"map":"map1","c-ms":["m1"],"c-per":"per1"},
    {"map":"map2","c-ms":{}}]
}
```
To retrieve all **m**s. [verify with topology](#retrieve)

```
> curl -i -X GET http://localhost:1202/ms
{"ms":["m1"]}
```
To retrieve all **p**s. [verify with topology](#retrieve)

```
> curl -i -X GET http://localhost:1202/ps
{"ps":["p1"]}
```
To retrieve **m**s contained by _**map1**_. [verify with topology](#retrieve)

```
> curl -i -X GET http://localhost:1202/maps/map1
{"ms":["m1"]}
```
To retrieve **p**s contained by _**per1**_. [verify with topology](#retrieve)

```
> curl -i -X GET http://localhost:1202/pers/per1
{"ps":["p1"]}
```
To retrieve **map** that _**m1**_ points to. [verify with topology](#retrieve)

```
> curl -i -X GET http://localhost:1202/ms/m1/maps
{"map":"map1"}
```
To retrieve **per** that _**map1**_ points to. [verify with topology](#retrieve)

```
> curl -i -X GET http://localhost:1202/maps/map1/pers
{"per":"per1"}
```
To retrieve **per** that _**p1**_ points to. [verify with topology](#retrieve)

```
> curl -i -X GET http://localhost:1202/ps/p1/pers
{"per":"per1"}
```
**To retrieve 1 or n random p that _m1_ can find(by m=>map=>per and p=>per)**. [verify with topology](#retrieve)

```
> curl -i -X GET http://localhost:1202/ms/m1/ps
{"ps":["p1"]}
> curl -i -X GET http://localhost:1202/ms/m1/ps?count=2
{"ps":["p1"]}
```
### DELETE
-----
To delete _**m1**_. [verify with topology](#retrieve)

```
> curl -i -X DELETE http://localhost:1202/ms/m1
```
To delete _**p1**_. [verify with topology](#retrieve)

```
> curl -i -X DELETE http://localhost:1202/ps/p1
```
To delete _**map1**_. [verify with topology](#retrieve)

```
> curl -i -X DELETE http://localhost:1202/maps/map1
```
To delete _**per1**_. [verify with topology](#retrieve)

```
> curl -i -X DELETE http://localhost:1202/pers/per1
```
To delete assignment of _**m1**_. [verify with topology](#retrieve)

```
> curl -i -X DELETE http://localhost:1202/ms/m1/maps
```
To delete assignment of _**p1**_. [verify with topology](#retrieve)

```
> curl -i -X DELETE http://localhost:1202/ps/p1/pers
```
To delete asignment of _**map1**_. [verify with topology](#retrieve)

```
> curl -i -X DELETE http://localhost:1202/maps/map2/pers
```
## What is next
Since **go-mapper** only supplies rest api, for the production environments of a distributed system, a node may be down before calling a **delete** rest api, and this scenario can be detected by a tcp connection or heartbeat packages, once connection interrupted, the data(maybe **p**) associated with the connection will be deleted, so a tcp implementation is within the plan.