**RISK**
Voici plein d'idées pour le jeu, on pourra définir ensuite quelles idées ici sont essentielles et lesquelles ne le sont pas.
- - -
# Le jeu

Notre projet est un jeu semblant légèrement à Risk où les joueurs doivent envahir des pays et bâtir un empire.
La carte du monde utilise un API de météo, la météo influe la partie rendant parfois certaines troupes plus faibles voir inutilisables, ou empêchant la création de ressources, ou encore forçant la retraite de toutes troupes présentes sur le territoire.
## Initialisation
Une carte de l'Europe est mise à disposition.
Un nouveau joueur peut commencer à jouer en envahissant un pays non-occupé. Il y place $N\_START$ humains.

En fonction du pays il va obtenir différentes ressources.

---
## Troupes et ressources
On peut imaginer qu'en plus de créer les troupes, on peut utiliser les ressources afin d'augmenter leur niveau  ( #extension ).
On peut être plus imaginatifs (dragons au lieu d'avions, centaurs au lieu de chevaux, bref si on veut changer la DA du jeu on peut).

J'ai noté des idées pour chaque troupe.
### Humains
La troupe de base coûtant le moins. Ils sont faibles face à toutes les autres unités mais le nombre fait la force!
### Chevaux
Fort contre les canons, faibles face aux avions.

Faiblesse(s) : Faibles quand il fait chaud
### Canons
Forts face aux avions.
Faiblesse(s) : Utilisation limitée ou Rouille après plusieurs pluies
### Avions
Forts contre les chevaux et les humains, faible face aux canons.
Faiblesse(s) : Inutilisables en tempête, perte de puissance sous la pluie

---
## Gameplay
Je pense qu'il faut un systeme de tours, limités à une fois par heure (temps modulable pour qd on voudra tester). Chaque heure on peut recevoir nos ressources, puis bouger nos troupes, puis attaquer, puis rebouger nos troupes et on doit attendre le prochain tour.
Chaque tour (donc une fois par heure), les conditions météorologiques sont mises à jour, des pays ou des troupes peuvent être perdues.

Après avoir attaqué et gagné un territoire d'un joueur adverse, les troupes du pays attaquant peuvent être déplacées pour continuer l'invasion (comme dans RISK).
On ne peut envahir un pays vide qu'une fois par tour. (TODO: définir comment ça se passe, peut-être qu'il faut payer des ressources pour obtenir certains pays).

### Déroulement d'un tour
Pendant une heure, les joueurs peuvent passer d'une phase à la suivante, dans l'ordre. Si on est pas passé à la dernière phase avant le prochain tour, on sera tout de même forcé à revenir à la première phase.

- Début de tour: la météo est actualisée et les ressources sont attribuées aux joueurs.
- Phase d'achat et déplacement: Le joueur peut acheter, placer et déplacer toutes les troupes
- Phase de combat: Le joueur ne peut plus déplacer ses troupes, il a le droit d'engager des guerres entre un de ses territoires et un territoire ennemi qui y est connecté. 
- Phase finale de déplacement: Le joueur peut replacer ses troupes afin d'optimiser ses défenses avant le prochain tour.
### Combat

Lorsqu'un territoire est attaqué, une guerre commence et le territoire ne peut pas être attaqué par un autre territoire avant la fin de la guerre et aucune troupe ne peut y être déplacé par le joueur en défense.
Dans RISK, il y a des avantages et des inconvénients différents pour l'attaquant et le défenseur (nombre de dés max + élevé chez l'attaquant, mais la défense gagne lorsqu'il y a égalité). (Peut-être que dans notre jeu, un avantage pour la défense peut-être que les conditions météorologique n'affectent que l'attaque, sachant que le mec qui attaque a déjà l'avantage de placer ses troupes pour attaquer).
Ici le combat se fera au hasard, avec des variables différentes en fonction du type de troupe et du nombre de troupes présentes ainsi que des phénomènes météorologiques. En gros on peut penser à un autobattleur. Peut-être qu'on limite le nombre de troupes qui peuvent se battre dans un combat, il faut donc parfois plusieurs combats pour en venir à bout d'un pays. Pour envahir un pays, il faut lui enlever sa dernière troupe.

L'attaquant peut stopper la guerre quand il veut, elle s'arrête automatiquement s'il envahis le pays.

Si le tour d'après commence pendant la guerre, la météo reste inchangée jusqu'à la fin de la guerre (je pense que c'est mieux). A la fin de la guerre le tour prochain commence alors instantanément. (il faudra timeout les guerres si jamais elles durent trop longtemps, sinon le pays défenseur ne pourra jamais utiliser son pays).


# Idées de développement

"L’application doit proposer du contenu générée par les utilisateurs, un bon exemple est l’integration d’une composante sociale (profils, commentaires, notes, messages, publications, . . . ).''
-> Ce n'est sûrement pas recommandé normalement, mais l'interaction produite par l'utilisateur seront:
	- L'envoi d'une destination et d'un JSON de troupes lors d'un déplacement.
	- L'envoi d'un jSON des troupes lors des début d'un combat pour une guerre.
(Sinon je ne sais pas trop... Et au pire si c'est trop complexe on ne fait pas de guerre mais juste tout le monde peut construire dans ses territoires, commenter le pays de l'autre et échanger des ressources. JSP si c'est moins complexe mais c'est sûrement plus utile de savoir coder ça pour de futurs projets).


# Ajouts Annexes

- Leaderboard
- Profiles joueurs configurables (Avec Pseudo, Titre, bannières, etc...) On pourrait acheter des éléments pour personnaliser son profil avec les ressources qu'on a débloqué en jeu.