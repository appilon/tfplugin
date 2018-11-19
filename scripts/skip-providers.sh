skip_provider () {
    # declare array of providers to skip
    declare -a arr=("")
    for i in "${arr[@]}"
    do
        if [ "$i" = "$1" ]
        then
            echo "skip"
            return
        fi
    done
}