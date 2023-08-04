1go build -tags base,viridian,hazelcastinternal,hazelcastinternaltest -ldflags "-s -w -X 'github.com/hazelcast/hazelcast-commandline-client/internal/viridian.EnableInternalOps=yes' -X 'github.com/hazelcast/hazelcast-commandline-client/clc/cmd.MainCommandShortHelp=Viridian Deploy Tool' -X 'github.com/hazelcast/hazelcast-commandline-client/internal.GitCommit=%GIT_COMMIT%' -X 'github.com/hazelcast/hazelcast-commandline-client/internal.Version=%CLC_VERSION%' -X 'github.com/hazelcast/hazelcast-go-client/internal.ClientType=CLC' -X 'github.com/hazelcast/hazelcast-go-client/internal.ClientVersion=%CLC_VERSION%'" -o build\vrd.exe ./cmd/clc