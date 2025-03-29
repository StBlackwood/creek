## **CreekDB - A Distributed Key-Value Store**
![Creek Logo](assets/creek_logo_med.png)

🚀 **CreekDB** is a high-performance, distributed key-value store designed for scalability and fault tolerance. It supports **replication, expiration (TTL), and eventual consistency** while maintaining a simple TCP-based interface.

---

## **🌟 Features**
- **🔑 Key-Value Storage:** Supports basic `SET`, `GET`, `DELETE` operations.
- **⏳ Expiry Support (TTL):** Allows keys to expire after a specified time.
- **📡 Replication:** Changes are propagated to peer nodes to ensure data consistency.
- **⚡ High Availability:** Designed for fault tolerance and scalability.
- **📜 Configurable Consistency Guarantees:** Future versions will allow tuning consistency vs. availability.
- **🛠️ Easy Integration:** Clients can interact with Creek via simple TCP commands.

---

## **🚀 Getting Started**

### **1️⃣ Clone the Repository**
```sh
git clone https://github.com/StBlackwood/creek.git
cd creek
```

### **2️⃣ Build and Run**
```sh
make build
```
Alternatively, for Windows users without _make_:

```sh
go build -o creek cmd/server/main.go
```

### **3️⃣ Start Multiple Nodes (WIP)**
Each node should have a unique **port** and a list of **peer nodes** for replication.

```sh
CREEK_CONF_FILE="node1.conf" ./creek
CREEK_CONF_FILE="node2.conf" ./creek
```

### **4️⃣ Connect via Python Cmdline Client**
```sh
python client/client_cmdline.py connect localhost:7690
```
- **Store a Key:** `SET user Alice`
- **Retrieve a Key:** `GET user`
- **Delete a Key:** `DELETE user`
- **Set Expiry (TTL):** `SET session abc123 5` (Expires in 5s)
- **Check TTL:** `TTL session`
- **Set Expiration:** `EXPIRE user 10`
- **Check Replication:** Run `GET user` on another node.

---

## **🔧 Configuration**
Creek can be customized via environment variables:

| **Variable**     | **Description**     | **Default**           |
|-----------------|---------------------|-----------------------|
| `CREEK_CONF_FILE` | Path to config file | `config/default.conf` |


Example:
```sh
export CREEK_CONF_FILE="config/dev.conf"
./creek
```

---

## **🛠️ Running Tests**
To ensure the Creek server works correctly, you can run the test suite.

### **Run Tests with Make**
```sh
make test
```

### **Run Tests with Go**
If you don’t have `make`, you can manually run tests using:
```sh
go test ./test
```

This will execute all test cases inside the `test` package.

---

## **🛠️ Architecture**
### **1️⃣ Data Storage**
- Uses an **in-memory key-value store** with optional TTL.
- Garbage collection periodically removes **expired keys**.

### **2️⃣ Replication**
- **Leaderless Replication:** Each node propagates updates to its peers.
- **Asynchronous Communication:** Non-blocking replication to avoid performance bottlenecks.

### **3️⃣ Fault Tolerance (Planned)**
- **Auto-Recovery:** If a node fails, surviving nodes continue to function.
- **Retry Mechanisms (Future Work):** Ensuring updates reach failed nodes upon recovery.

### **4️⃣ Configurable Consistency Guarantees (Planned)**
- **Eventual Consistency:** Ensures all nodes eventually converge.
- **Strong Consistency (Future):** Using **quorum-based reads/writes**.
- **Partitioning (Future):** Distributing keys across multiple nodes.

---

## **📌 Roadmap**
✔ **Basic Key-Value Store**  
✔ **Garbage Collection for Expired Keys**  
✔ **Persistent Storage through commit logs**  
✔ **Crash Recovery**  
✔ **Replication Across Nodes**  
🔜 **Configurable Consistency Levels**  
🔜 **Basic Fault Tolerance**  
🔜 **Automatic Data Partitioning**  
🔜 **Distributed Transactions**

[Next Roadmap tasks are outlined in a trello board here](https://trello.com/b/p2PbyoZV)

---

## **👨‍💻 Contributing**
We welcome contributions! Please follow these steps:
1. **Fork the repo** and clone it locally.
2. Create a **feature branch**.
3. Write **tests** for new functionality.
4. Submit a **pull request**.

---

## **📜 License**
Creek is open-source and available under the **MIT License**.
