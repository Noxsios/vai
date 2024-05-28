Output demo.gif

Require "yq"

Set Shell "bash"
Set FontSize 32
Set Width 1600
Set Height 1200

Hide
Type "alias vai=bin/vai" Enter

Type "cat <<EOF > vai.yaml" Enter
Type "echo:" Enter
Type '  - run: echo "\$MESSAGE"' Enter
Type "    with:" Enter
Type "      message: \${{ input }}" Enter
Type "" Enter
Type "remote-echo:" Enter
Type "  - uses: pkg:github/noxsios/vai@main?task=echo#testdata/simple.yaml" Enter
Type "    with:" Enter
Type "      message: hello from main" Enter
Type "EOF" Enter
Ctrl+L
Show

Type "yq vai.yaml" Sleep 500ms Enter

Enter

Sleep 3s

Type "vai echo --with message='Hello World!'" Sleep 500ms  Enter

Sleep 3s

Type "vai remote-echo" Sleep 500ms  Enter

Sleep 5s

Hide
Type "rm vai.yaml" Enter
Show
