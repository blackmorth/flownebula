<?php

function a() {
    usleep(20000);
    b();
    b();
}

function b() {
    usleep(10000);
    c();
}

function c() {
    usleep(5000);
}

a();