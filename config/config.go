package config

import (
    "fmt"
    "log"
    "strconv"
    "strings"

    "github.com/albertofem/gloster/node"
)

type Config struct {
    Addr    node.Node
    Cluster map[string]node.Node
    ElectionTimeout int // ms
    DatabasePath string
}

func NewConfig(addr, cluster string, database string) *Config {
    addrNode, err := ParseNodeConfig(addr)
    if err != nil {
        log.Println("Error parsing Local Address ", addr)

        return nil
    }

    return &Config{
        Addr:    addrNode,
        Cluster: ParseClusterConfig(cluster),
        ElectionTimeout: 1000,
        DatabasePath: database,
    }
}

func ParseClusterConfig(clusterList string) map[string]n.Node {
    parts := strings.Split(clusterList, ",")
    r := make(map[string]node.Node, len(parts))

    if len(parts) > 0 && len(parts[0]) > 0 {
        for _, nodePart := range parts {
            node, err := ParseNodeConfig(nodePart)

            if err != nil {
                continue
            }

            r[node.String()] = node
        }

        return r
    }

    return nil
}

func ParseNodeConfig(node string) (node.Node, error) {
    nodeParts := clear(node)

    if len(nodeParts) != 0 {
        node = nodeParts[0]
    }

    p := strings.Split(node, ":")

    if len(p) != 2 {
        return node.Node{}, fmt.Errorf("Error building Node on config: %s", p)
    }

    port, err := strconv.ParseInt(p[1], 10, 64)

    if err != nil {
        return node.Node{}, err
    }

    return node.Node{Host: p[0], Port: int(port)}, nil
}

func clear(node string) []string {
    return strings.Split(node, "\n")
}
