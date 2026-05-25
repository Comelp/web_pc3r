
# RAIN WAR
> Lien vers l'application : [https://rain-war.lance-perlick.synology.me/](https://rain-war.lance-perlick.synology.me/)

---

## 1. Présentation générale du projet

L'application qu'on a développée est un jeu web multijoueur de conquête territoriale à l'échelle de l'Europe. L'idée de base c'est que chaque joueur peut s'inscrire, puis essayer de prendre le contrôle de différents pays européens en achetant des troupes, en attaquant d'autres joueurs, ou en se défendant.

Les principales fonctionnalités du jeu sont :
- **Inscription et connexion** avec un système de sessions basique.
- **Une carte interactive** de l'Europe où on voit en temps réel qui contrôle quoi.
- **Achat et déploiement de troupes** : soldats, tanks, avions.
- **Attaque et défense de pays**, avec un système de combat résolu automatiquement à la fin de chaque phase.
- **Amélioration des pays** pour augmenter leur production d'or.
- **Affichage de la météo** pour chaque pays, récupérée depuis une API externe.
- **Un système de phases cycliques** (toutes les 10 minutes) avec une phase "Attaque" et une phase "Distribution".

Le jeu tourne sur un serveur Go qui gère toute la logique, et le client est une SPA React. Les règles du jeu sont accessibles directement depuis le site.

---

## 2. API Web utilisée

Pour enrichir l'expérience de jeu, on a intégré **WeatherAPI** (`api.weatherapi.com`), une API météo publique et gratuite (dans une certaine limite de requêtes).

### Ce que l'API fournit

On récupère deux informations pour chaque pays :
- `current.temp_c` : la température actuelle en degrés Celsius.
- `current.condition.text` : une description textuelle de la météo (ex. "Sunny", "Rain", "Thunderstorm", etc.).

### Comment on l'utilise

Le serveur appelle cette API dans le fichier `weatherpull.go`. Comme l'API prend un nom de ville en paramètre et pas un nom de pays, on a fait un mapping manuel pour certains pays (par exemple `"Czechia"` devient `"Prague"`).

L'URL appelée ressemble à ça :

```
http://api.weatherapi.com/v1/current.json?key=XXXX&q=Paris
```

Les données météo sont stockées dans le champ `Meteo` de chaque `CountryInfo` en mémoire, puis sauvegardées dans `server/data/countryInfos.json`. Côté client, ces infos s'affichent dans le panneau d'information d'un pays quand on clique dessus.

La météo a aussi un effet sur la logique de jeu : si la météo d'un pays contient "rain" ou "thunder", le leader peut perdre son contrôle du pays à la fin de la phase.

### Quand l'API est appelée

La fonction `StartWeatherPuller()` tourne en arrière-plan côté serveur et appelle l'API toutes les **10 minutes** via un ticker Go pour tous les pays listés.

> **Note** : la clé API est actuellement codée en dur dans `weatherpull.go`. Pour une mise en production sérieuse, il faudrait la mettre dans une variable d'environnement.

---

## 3. Fonctionnalités de l'application

### Côté serveur (Go)

- Gestion de l'état du jeu (`GameState`) en mémoire, protégé par un `sync.RWMutex` pour éviter les conflits d'accès concurrent.
- Endpoints HTTP pour toutes les actions : authentification, récupération d'état, actions de jeu.
- Deux processus périodiques en arrière-plan : la mise à jour météo et le gestionnaire de phases.
- Persistance des données dans des fichiers JSON sous `server/data/`.

### Côté client (React)

- Carte interactive de l'Europe avec coloration des pays par joueur.
- Icônes `⚔️` et `🏛️` affichées sur les pays en combat ou en conquête.
- Panneau utilisateur avec l'or disponible et les boutons d'achat de troupes.
- Formulaire de déploiement de troupes (`DeployForm`).
- Timer de phase affiché en haut de page.
- Popup récapitulatif de fin de phase avec les résultats des combats, les conquêtes, l'or gagné, etc.

---

## 4. Cas d'utilisation

### Yvan profite d'une phase de paix
Yvan se reconnecte après deux jours d'absence. Il voit que la phase en cours est "Paix 🤝" et que son compteur d'or a bien grimpé, ses pays ont produit pendant ce temps. Il possède la France et l'Espagne, toutes les deux au niveau 1. Il décide de les améliorer au niveau maximum en cliquant dessus l'une après l'autre et en confirmant l'upgrade à chaque fois. Ça lui coûte une bonne partie de son or mais la production future sera bien meilleure.
Il jette ensuite un œil à l'Italie. Le pays est libre, son ancien propriétaire dont le pseudo est MichaelJackson69 l'a perdu la semaine dernière à cause d'une tempête détectée par la météo. Yvan clique sur l'Italie et appuie sur "Conquérir". Il n'a plus qu'à attendre la fin de la phase.

## Alice découvre le jeu
Alice tombe sur le lien du jeu partagé par un ami. Elle arrive sur la page et voit une carte de l'Europe couverte de couleurs vives : du rouge sur la France, du bleu sur l'Allemagne, du vert un peu partout. Elle ne comprend pas vraiment ce qui se passe alors elle clique sur "Règles". Après avoir lu, elle comprend le principe : conquérir des pays, acheter des troupes, survivre aux autres joueurs. Elle crée un compte avec la couleur orange, et découvre qu'elle commence avec un peu d'or mais aucun pays. Elle clique sur la Pologne — personne ne la contrôle. Le bouton "Conquérir" s'affiche. Elle clique, et attend la fin de la phase pour voir si ça marche.

## Karim tente une attaque risquée
Karim est en phase "Attaque 🪖" et il lorgne sur l'Allemagne depuis un moment, contrôlée par Alice qui a peu de troupes en défense d'après ce qu'il a vu au dernier tour. Il achète trois tanks, les déploie en attaque sur l'Allemagne, et appuie sur "Attaquer". Le timer de phase indique encore quatre minutes. Il attend, un peu stressé. La fin de phase arrive, le popup s'affiche : "Attaque ratée, Karim a échoué contre l'Allemagne (owner: Alice)". Alice avait apparemment déployé des soldats en défense juste avant la fin. Karim perd ses tanks et repart sans rien. La prochaine fois il vérifiera la météo d'abord.

---

## 5. Base de données

On utilise des **fichiers JSON** stockés dans `server/data/`.

### Fichiers persistés

| Fichier | Contenu |
|---|---|
| `countryInfos.json` | État de tous les pays |
| `playerInfos.json` | Données de tous les joueurs |
| `phaseSnapshot.json` | Snapshot des pays en début de phase |
| `phasePopup.json` | Popup récapitulatif de fin de phase |

### Schémas

```
CountryInfo
- leader_id         : string | null
- attacked_by       : string | null
- conquered_by      : string | null
- meteo             : { temperature: float, condition: string } | null
- produced_gold     : int
- level             : int  (0 à 3)
- troops_attacking  : { type: string, count: int }
- troops_defending  : { type: string, count: int }
```

```
PlayerInfo
- couleur   : string
- password  : string
- gold      : int
- troops    : { soldiers: int, tanks: int, planes: int }
```

Les sessions sont gardées en mémoire uniquement (non persistées), donc elles disparaissent si le serveur redémarre.

---

## 6. Mise à jour des données

### Actions utilisateur

Quand un joueur fait une action (achat, déploiement, attaque, upgrade...), le handler correspondant modifie l'état en mémoire (`GameState`) et réécrit les fichiers JSON concernés (`playerInfos.json`, `countryInfos.json`).

### Processus périodiques

Deux goroutines tournent en permanence en arrière-plan :

- **`StartWeatherPuller()`** : appelle l'API météo toutes les 10 minutes et met à jour les champs `Meteo` de chaque pays, puis persiste `countryInfos.json`.
- **`StartPhaseWatcher()`** : toutes les 10 minutes, il sauvegarde un snapshot (`phaseSnapshot.json`), attend la fin de phase, puis appelle `onPhaseEnd()` qui :
  - résout les combats,
  - distribue l'or,
  - applique les améliorations de pays,
  - persiste `countryInfos.json` et `playerInfos.json`,
  - génère `phasePopup.json`.

### Synchronisation côté client

Le client ne reçoit pas de notifications push — il interroge le serveur périodiquement :
- `/getMapInfos` toutes les **5 secondes** (dans `GameMap`).
- `/getPhasePopup` toutes les **6 secondes** (dans `GameMap`).
- `/getPhase` toutes les **5 secondes** (dans `Timer`).

Les modifications sensibles (résolution de combats) sont atomiques grâce aux méthodes `SetCountries` et `SetPlayers` qui prennent le lock du `GameState`.

---

## 7. Description du serveur

### Architecture choisie

Le serveur est en **Go** et expose des endpoints HTTP. L'architecture est un mélange entre ressources et services — c'est un choix pragmatique pour un jeu.

D'un côté, on a des endpoints de lecture qui exposent les ressources (`/getMapInfos`, `/me`, `/getState`). De l'autre, on a des endpoints actionnels qui déclenchent des opérations métier (`/buyTroop`, `/attackCountry`, `/deployTroop`, etc.).

### Fichiers principaux

| Fichier | Rôle |
|---|---|
| `server.go` | Point d'entrée, routage HTTP, `GameState`, démarrage des goroutines |
| `api.go` | Handlers HTTP (auth + actions de jeu), gestion des sessions |
| `phase_manager.go` | Logique de fin de phase, résolution des combats, popups |
| `weatherpull.go` | Appels à WeatherAPI, mise à jour météo périodique |

### Particularités

- **Concurrence** : `sync.RWMutex` sur `GameState` pour éviter les race conditions entre les goroutines et les requêtes HTTP.
- **Sessions** : stockées en mémoire dans une map avec un mutex dédié. Le client envoie un cookie `session_id` avec chaque requête.
- **Persistance** : réécriture complète des fichiers JSON à chaque modification (pas de SGBD).

---

## 8. Description du client

### Type d'application

C'est une **Single Page Application (SPA)** en React. Le serveur Go sert les fichiers statiques compilés (dans `dist/`) via `http.FileServer`. Il n'y a pas de navigation multi-pages — tout se passe sur une seule page sans rechargement.

### Écrans principaux

- **Page de login/inscription** : composant `Login`, permet de se connecter ou de créer un compte.
- **Écran principal** : composant `GameMap` (la carte interactive) + `UserInterface` (panneau d'achat) + `Timer` (phase et countdown).
- **Popup de fin de phase** : rendu par `GameMap.renderPhasePopup()`, affiché par-dessus la carte.
- **Panneau d'info d'un pays** : affiché en bas à gauche quand on clique sur un pays, avec les infos et les boutons d'action.
- **Formulaire de déploiement** : composant `DeployForm`, s'affiche à la place du panneau d'info quand on veut déployer des troupes.

### Où sont faits les appels au serveur

| Composant | Endpoints appelés |
|---|---|
| `AuthContext` | `/me`, `/login`, `/register`, `/logout`, `/buyTroop` |
| `GameMap` | `/getMapInfos`, `/getState`, `/getPhasePopup`, `/ackPhasePopup`, `/attackCountry`, `/conquerCountry`, `/upgradeCountry`, `/retraiteTroop` |
| `DeployForm` | `/deployTroop` |
| `Timer` | `/getPhase` |

---

## 9. Requêtes et réponses HTTP

### Récupérer l'utilisateur courant

```http
GET /me
Cookie: session_id=abc123
```

```json
{
  "username": "jean",
  "troops": {
    "soldiers": 5,
    "tanks": 2,
    "planes": 0
  },
  "gold": 3000
}
```

---

### Acheter une troupe

```http
POST /buyTroop?troop=soldiers
Cookie: session_id=abc123
```

```json
{
  "gold_remaining": 2000,
  "troops": {
    "soldiers": 6,
    "tanks": 2,
    "planes": 0
  }
}
```

---

### Déployer des troupes en attaque

```http
POST /deployTroop?country=France&mode=attacking&troop=soldiers&count=3
Cookie: session_id=abc123
```

Réponse : texte brut indiquant succès ou erreur (ex. `"OK"` ou `"Not enough troops"`).

---

### Récupérer les infos de la carte

```http
GET /getMapInfos
```

```json
{
  "France": {
    "color": "#e74c3c",
    "is_attacked": false,
    "is_conquered": false,
    "attacked_by": null,
    "conquered_by": null,
    "leader_id": "jean",
    "level": 1,
    "meteo": {
      "temperature": 14.5,
      "condition": "Partly cloudy"
    }
  },
  "Germany": {
    "color": "#3498db",
    "is_attacked": true,
    "is_conquered": false,
    "attacked_by": "pierre",
    "leader_id": "marie",
    ...
  }
}
```

---

### Popup de fin de phase

```http
GET /getPhasePopup
Cookie: session_id=abc123
```

```json
{
  "phase": "Attaque 🪖",
  "conquered": [
    { "country": "Poland", "by": "pierre" }
  ],
  "gained_attack": [
    { "attacker": "jean", "country": "Spain", "previous_owner": "marie" }
  ],
  "lost_attack": [],
  "lost_to_weather": [
    { "country": "Norway", "previous_owner": "pierre", "reason": "Heavy rain" }
  ],
  "improved": [],
  "gold_earned": [
    { "player": "jean", "amount": 5000 }
  ]
}
```

---

### Acquitter le popup

```http
POST /ackPhasePopup
Cookie: session_id=abc123
```

---

## 10. Choix d'architecture

### SPA vs application classique

On a choisi une **SPA React** pour avoir une interface dynamique qui se met à jour sans rechargement de page. Ça correspond bien à un jeu où les données changent fréquemment et où l'utilisateur doit voir les mises à jour en presque temps réel.

### Approche services/ressources

Comme mentionné plus haut, on a un mélange des deux. Certaines opérations sont clairement des "commandes" (`/attackCountry`, `/buyTroop`) et d'autres sont des lectures d'état (`/getMapInfos`, `/me`). Essayer de tout forcer dans un modèle REST pur aurait compliqué les choses sans vraiment apporter de bénéfice.

### Pas de framework CSS

On n'a pas utilisé de framework CSS, les styles sont essentiellement en inline CSS dans les composants React. C'est pas idéal pour la maintenabilité, mais étant pas à l'aise avec le formattage, nous avons demandé à l'IA Claude de nous aider et c'est la solution qu'elle a proposé. 

---

## 11. Description des composants serveur

### `MeHandler` — `GET /me`

Retourne les informations du joueur actuellement connecté à partir du cookie `session_id`.

- **GET** : renvoie `{ username, troops, gold }` si la session est valide ; erreur 4xx sinon.

---

### `RegisterHandler` — `POST /register`

Crée un nouveau joueur.

- **POST** : paramètres `username`, `password`, `couleur` via le formulaire. Crée un `PlayerInfo` avec des ressources initiales et sauvegarde dans `playerInfos.json`.

---

### `LoginHandler` — `GET /login` et `POST /login`

Authentifie un joueur.

- **GET** : affiche la page de login (dans notre cas, c'est le client React qui gère ça).
- **POST** : vérifie les identifiants, crée un cookie `session_id` via `newSessionID(username)` et stocke la session en mémoire.

---

### `LogoutHandler` — `POST /logout`

Déconnecte le joueur.

- **POST** : supprime la session côté serveur et expire le cookie.

---

### `MapInfosHandler` — `GET /getMapInfos`

Fournit l'état de tous les pays pour l'affichage de la carte.

- **GET** : renvoie une map JSON des `CountryMapInfo` (couleur, leader, statut d'attaque/conquête, météo, niveau...).

---

### `StateHandler` — `GET /getState`

Retourne les informations détaillées d'un pays spécifique.

- **GET** : paramètre `country` en query string ; renvoie un `CountryInfo` complet.

---

### `PhaseHandler` — `GET /getPhase`

Retourne la phase actuelle du jeu.

- **GET** : renvoie `{ "phase": "Attaque 🪖" }` ou `{ "phase": "Paix 🤝" }`.

---

### `PhasePopupHandler` — `GET /getPhasePopup`

Retourne le popup récapitulatif de fin de phase.

- **GET** : renvoie la structure du popup ou `null` s'il n'y en a pas.

---

### `AckPhasePopupHandler` — `POST /ackPhasePopup`

Acquitte le popup de fin de phase.

- **POST** : marque le popup comme lu côté serveur.

---

### `BuyTroopHandler` — `POST /buyTroop`

Achète une troupe pour le joueur courant.

- **POST** : paramètre `troop` en query string (`soldiers`, `tanks` ou `planes`). Déduit l'or et ajoute la troupe. Renvoie l'état mis à jour du joueur.

---

### `DeployTroopHandler` — `POST /deployTroop`

Déploie des troupes sur un pays.

- **POST** : paramètres `country`, `mode` (`attacking` ou `defending`), `troop`, `count`. Déplace les troupes depuis le stock du joueur vers le slot correspondant du pays.

---

### `AttackHandler` — `POST /attackCountry`

Lance une attaque sur un pays.

- **POST** : paramètre `country`. Marque le pays comme étant attaqué (`attacked_by`).

---

### `ConquerHandler` — `POST /conquerCountry`

Lance une conquête sur un pays non occupé.

- **POST** : paramètre `country`. Marque le pays comme étant en cours de conquête (`conquered_by`).

---

### `UpgradeHandler` — `POST /upgradeCountry`

Améliore un pays dont le joueur est leader.

- **POST** : paramètre `country`. Vérifie que le joueur est leader, que la phase est correcte et qu'il a assez d'or. Met à jour `CountryInfo.level`.

---

### `RetraiteTroopHandler` — `POST /retraiteTroop`

Retire des troupes déployées sur un pays.

- **POST** : paramètres `country` et `mode`. Rend les troupes au joueur.

---

## 12. Technologies clientes

### React

- **`AuthContext`** : un Context Provider React qui garde en mémoire l'utilisateur connecté, son or et ses troupes. Il expose `refreshAuth()` (pour recharger les infos depuis `/me`) et `buyTroop()`. Tous les composants qui ont besoin de ces données y accèdent via `static contextType = AuthContext`.
- **`GameMap`** : composant de classe principal. Gère l'affichage de la carte SVG, les clics sur les pays, les timers de polling et tout ce qui touche aux interactions de jeu.
- **`UserInterface`** : panneau latéral qui affiche l'or et les boutons d'achat.
- **`DeployForm`** : formulaire modal pour déployer les troupes.
- **`Timer`** : affiche la phase en cours et un countdown.

### Fetch API

Tous les appels réseau sont faits avec l'API `fetch` native du navigateur. Les appels qui nécessitent l'authentification incluent `credentials: 'include'` pour envoyer le cookie de session.

---

## 13. Plan de la partie client

### Écrans et navigation

L'application n'a qu'un seul écran principal qui change d'affichage selon l'état de l'authentification (login ou carte de jeu). Il y a deux autres écrans simples: un HTML qui donne les règles du jeu, et un écran d'authentification.


### Flux et événements importants

**Au chargement :**
- `AuthContext` appelle `/me` pour récupérer l'utilisateur courant.
- `GameMap.componentDidMount()` appelle `/getMapInfos` immédiatement, puis démarre deux intervalles :
  - `setInterval(() => this.fetchMapInfos(), 5000)` pour rafraîchir la carte.
  - `setInterval(() => this.fetchPhasePopup(), 6000)` pour les popups.

**Clic sur un pays :**
- `handleCountryClick` détecte l'élément SVG cliqué via `closest('[data-country]')`.
- Appel `GET /getState?country=...` pour récupérer les infos détaillées.
- Le pays sélectionné est mis en surbrillance (couleur violette).
- Le panneau d'info s'affiche en bas à gauche.

**Achat de troupes :**
- `UserInterface` → `AuthContext.buyTroop()` → `POST /buyTroop?troop=...`
- Au retour, `AuthContext` met à jour `playerGold` et `troops` → les composants React se re-rendent automatiquement.

**Déploiement :**
- Clic sur "Déployer la Défense" / "Déployer l'Attaque" → affiche `DeployForm`.
- `DeployForm` envoie `POST /deployTroop`.
- Après succès : `refreshAuth()` + `GET /getState` pour mettre à jour l'UI.

**Fin de phase :**
- Le polling de `fetchPhasePopup()` détecte un popup non vide.
- `renderPhasePopup()` affiche la popup par-dessus tout le reste.
- Fermeture → `POST /ackPhasePopup` + `fetchMapInfos()`.

**Coloration de la carte :**
- Après chaque appel `/getMapInfos`, `applyCountriesColor()` parcourt tous les pays et applique `el.style.fill = info.color` sur les éléments SVG correspondants.
- Les pays en attaque/conquête ont une icône emoji (`⚔️` / `🏛️`) positionnée en absolu par-dessus la carte.

---

## 14. Liste des appels AJAX

| Méthode | URL | Rôle | Composant |
|---|---|---|---|
| `GET` | `/me` | Récupérer l'utilisateur courant | `AuthContext` |
| `POST` | `/login` | Connexion | `Login` |
| `POST` | `/register` | Inscription | `Login` |
| `POST` | `/logout` | Déconnexion | `Login` |
| `GET` | `/getMapInfos` | État des pays pour la carte | `GameMap` |
| `GET` | `/getState?country=...` | Infos détaillées d'un pays | `GameMap` |
| `GET` | `/getPhase` | Phase actuelle | `GameMap`, `Timer` |
| `GET` | `/getPhasePopup` | Popup de fin de phase | `GameMap` |
| `POST` | `/ackPhasePopup` | Acquitter le popup | `GameMap` |
| `POST` | `/buyTroop?troop=...` | Acheter une troupe | `AuthContext` |
| `POST` | `/deployTroop?country=...&mode=...&troop=...&count=...` | Déployer des troupes | `DeployForm` |
| `POST` | `/attackCountry?country=...` | Initier une attaque | `GameMap` |
| `POST` | `/conquerCountry?country=...` | Initier une conquête | `GameMap` |
| `POST` | `/upgradeCountry?country=...` | Améliorer un pays | `GameMap` |
| `POST` | `/retraiteTroop?country=...&mode=...` | Retirer des troupes | `GameMap` |

Tous les appels authentifiés incluent `credentials: 'include'` pour envoyer le cookie `session_id`.

---

## 15. Utilisation de l'IA

On a utilisé l'IA (Claude) exclusivement pour la **partie client**. Plus précisément, ça nous a aidé sur deux points où on était vraiment peu à l'aise :

- Le **code JavaScript** et la logique React (gestion des états, lifecycle components, callbacks enchaînés...).
- Le **formatage** du rapport et la mise en forme Markdown.

La partie serveur en Go a été écrite sans aide de l'IA.
