# FlowNebula

Local PHP profiler that generates call graphs.

## Build extension

cd core

phpize
./configure --enable-flownebula
make

## Run

php -dextension=modules/flownebula.so examples/test.php

Trace file:

/tmp/nebula.trace

## Build graph

php analyzer/build_graph.php /tmp/nebula.trace nebula.json

## View graph

Open viewer/index.html in browser


## Docker (Linux, macOS, Windows)

You can run FlowNebula entirely inside Docker, regardless of your host OS.

### Build the image

```bash
docker build -t flownebula .
```

Or with docker compose:

```bash
docker compose build
```

### Run a profiled script

By default the image already has the extension compiled and enabled.
You can run the example script like this:

```bash
docker compose run --rm flownebula php examples/test.php
```

This will generate a trace file inside the container at `/tmp/nebula.trace`
(also mounted to `./data/nebula.trace` on the host by default).

### Build the graph inside Docker

Generate `nebula.json` from the trace:

```bash
docker compose run --rm flownebula \
  php analyzer/build_graph.php /tmp/nebula.trace /tmp/nebula.json
```

The resulting JSON will be available on the host at `./data/nebula.json`.

### View the graph

Start the HTTP server that serves the viewer:

```bash
docker compose up flownebula
```

Then open in your browser:

```
http://localhost:8080
```

The viewer will load `nebula.json` from the container (backed by the `./data`
directory on the host).




# Tester FlowNebula sous Windows

Ce guide explique comment compiler et tester **FlowNebula** sur Windows.
L’objectif est d’obtenir une extension PHP (`flownebula.dll`), exécuter un script PHP avec le profiler activé, puis visualiser le graphe d’appels.

---

# 1. Pré-requis

Installer les outils suivants :

* PHP installé sur Windows
* Visual Studio (version compatible avec PHP)
* Git
* Un terminal (PowerShell ou cmd)

Télécharger les **PHP SDK tools** utilisés pour compiler les extensions :

https://github.com/php/php-sdk-binary-tools

Extraire par exemple dans :

```
C:\php-sdk
```

---

# 2. Préparer l’environnement de build

Ouvrir un **Developer Command Prompt for VS**.

Puis initialiser l’environnement PHP :

```
cd C:\php-sdk
phpsdk-vs17-x64.bat
```

Créer un dossier de travail :

```
mkdir C:\php-sdk\build
cd C:\php-sdk\build
```

---

# 3. Télécharger les sources PHP

Cloner les sources correspondant à la version de PHP installée :

```
git clone https://github.com/php/php-src.git
cd php-src
```

Puis initialiser les dépendances :

```
phpsdk_deps --update
```

---

# 4. Ajouter l’extension FlowNebula

Créer le dossier :

```
ext\flownebula
```

Copier dedans le contenu de ton dossier `core/` :

```
config.m4
php_flownebula.h
flownebula.c
```

La structure doit être :

```
php-src
 └── ext
     └── flownebula
         ├── config.m4
         ├── php_flownebula.h
         └── flownebula.c
```

---

# 5. Générer le système de build

Dans le dossier `php-src` :

```
buildconf
```

Puis configurer PHP avec l’extension :

```
configure --enable-flownebula
```

---

# 6. Compiler l’extension

Compiler PHP (ou au minimum l’extension) :

```
nmake
```

Si tout se passe bien, l’extension sera générée ici :

```
php-src\x64\Release\php_flownebula.dll
```

---

# 7. Charger l’extension dans PHP

Copier le fichier `.dll` dans le dossier `ext` de PHP.

Exemple :

```
C:\php\ext\php_flownebula.dll
```

Modifier `php.ini` :

```
extension=php_flownebula.dll
```

Vérifier que l’extension est chargée :

```
php -m
```

Tu dois voir :

```
flownebula
```

---

# 8. Tester le profiler

Créer un script de test :

```
examples/test.php
```

Exemple :

```php
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
```

Lancer le script :

```
php examples/test.php
```

Si tout fonctionne, un fichier de trace sera généré :

```
C:\tmp\nebula.trace
```

---

# 9. Construire le graphe

Utiliser l’analyseur :

```
php analyzer/build_graph.php C:\tmp\nebula.trace nebula.json
```

Un fichier sera généré :

```
nebula.json
```

---

# 10. Visualiser la nébuleuse

Ouvrir dans un navigateur :

```
viewer/index.html
```

La page affichera un graphe dynamique :

* chaque **nœud** représente une fonction
* chaque **lien** représente un appel
* l’épaisseur du lien correspond au nombre d’appels

---

# 11. Workflow complet

```
php examples/test.php
php analyzer/build_graph.php C:\tmp\nebula.trace nebula.json
ouvrir viewer/index.html
```

---

# 12. Dépannage

### L’extension ne se charge pas

Vérifier :

```
php -i | findstr flownebula
```

### PHP ne trouve pas la DLL

Ajouter le dossier `ext` dans `PATH` ou vérifier `php.ini`.

### Aucun fichier trace

Vérifier que le dossier de sortie existe :

```
C:\tmp
```

---

# 13. Étapes suivantes

Une fois la capture validée, FlowNebula pourra évoluer vers :

* capture des fonctions internes PHP
* capture mémoire
* flamegraph
* visualisation temps réel

À ce stade, tu disposes déjà d’un **profiler PHP local fonctionnel**.
