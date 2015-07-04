#!/bin/bash

mkdir -p build

targets=(
	darwin_amd64
	linux_amd64
	linux_arm
	linux_arm64
	windows_amd64
)

cat <<EOF > build/index.html
<!doctype html>
<html>
<head>
	<title>sith downloads</title>
</head>

<body>
EOF

for target in ${targets[@]}
do
	os=${target%_*}
	arch=${target#*_}
	if [ "$os" = "windows" ]
	then
		exe=sith_$target.exe
	else
		exe=sith_$target
	fi
	#GOOS=$os GOARCH=$arch go build -o build/$exe
	echo '<a href="'${exe}'">'${exe}'</a><br>'
done >> build/index.html

cat <<EOF >> build/index.html
</body>

</html>
EOF



