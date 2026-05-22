import React from 'react';
import { createRoot } from 'react-dom/client';
import GameMap from './gameMap';
import Timer from './timer';
import Login from './login';
import Rules from "./rules";

function App() {
  return (
  <div>
    <div>
      <h1>Bienvenue dans le jeu de conquête de l'Europe !</h1>
    </div>
    <Rules />
    <Login />
    <Timer />
    <GameMap />
  </div>
  );
}

const root = createRoot(document.getElementById('root'));
root.render(<App />);