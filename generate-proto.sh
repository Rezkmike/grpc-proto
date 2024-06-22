docker run --rm --platform linux/arm64 -v $(pwd):/workspace -w /workspace rvolosatovs/protoc \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  /workspace/proto/example.proto

mkdir -p server/generated
mkdir -p client/generated

cp -r generated/* server/generated/
cp -r generated/* client/generated/

