<?php

function foo() {
    usleep(10000);
}

foo();

function bar($n) {
    if ($n <= 0) return;
    bar($n - 1);
}
bar(3);
echo "OK\n";