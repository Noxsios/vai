Output demo.gif

Require "yq"

Set Shell "bash"
Set FontSize 32
Set Width 1600
Set Height 1200
Set FontFamily "Zed Mono Extended"

Hide
Type "alias vai=bin/vai" Enter

Type "cat <<EOF > vai.yaml" Enter
Type "echo:" Enter
Type '  - run: echo "\$MESSAGE"' Enter
Type "    with:" Enter
Type "      message: input" Enter
Type "" Enter
Type "remote-echo:" Enter
Type "  - uses: pkg:github/noxsios/vai@main?task=echo#testdata/simple.yaml" Enter
Type "    with:" Enter
Type "      message: "
Type '"'
Type "'hello from main'
Type '"' Enter
Type "EOF" Enter
Ctrl+L
Show

Type "yq vai.yaml" Sleep 500ms Enter

Enter

Sleep 3s

Type "vai echo --with message='Hello World!'" Sleep 500ms  Enter

Sleep 5s

Type "vai remote-echo -l debug" Sleep 500ms  Enter

Sleep 7s

Hide
Type "git restore vai.yaml" Enter
Show
