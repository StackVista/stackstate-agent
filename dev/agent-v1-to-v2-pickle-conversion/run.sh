pyv="$(python2.7 -V 2>&1)"

if [ "$pyv" != "Python 2.7.18" ]; then
    echo "Failure ...."
    echo ""
    echo "Incorrect python version detected, Detected $pyv and the required version is Python 2.7.18"
    exit 0
else
    pip install -r requirements.txt

    echo ""
    echo ""
    echo "Executing the pickle conversion process ..."
    echo ""

    python2.7 main.py
fi
