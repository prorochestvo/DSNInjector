# DSNInjector
DSNInjector is a Go library for managing and parsing Data Source Names (DSNs). 
It simplifies the handling of connection strings, environment variables, and configuration for various data sources, such as databases or other services requiring structured connection information.



## Installation

```bash
go get -u github.com/prorochestvo/dsninjector
```


## Usage

```go
cnf := "mysql://user:password@localhost:3306/dbname?charset=utf8"

dns, err := dsninjector.Parse(dsn)
if err != nil {
    panic(err)
}

// access parsed data
fmt.Println(dns.Driver())          // "mysql"
fmt.Println(dns.Host())            // "localhost"
fmt.Println(dns.Port())            // 3306
fmt.Println(dns.Database())        // "dbname"
fmt.Println(dns.Option("charset")) // "utf8"
```



## Contributing

Contributions are welcome! Please submit issues or pull requests on the GitHub repository.



## License

This project is licensed under the [MIT License](LICENSE).
