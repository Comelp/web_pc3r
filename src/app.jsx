import React from 'react';
import { createRoot } from 'react-dom/client';
import GameMap from './gameMap';

function App() {
  return (
  <div>
    <div>
      <h1>Bienvenue dans le jeu de conquête de l'Europe !</h1>
    </div>
    <GameMap />
  </div>
  );
}

const root = createRoot(document.getElementById('root'));
root.render(<App />);