<?php
putenv('FLOWNEBULA_AGENT_ADDR=127.0.0.1:8135'); // Remplace par l'adresse de ton agent Go

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