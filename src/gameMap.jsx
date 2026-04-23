import React, {Component} from 'react';
import EuropeMap from '../assets/europeMap.svg';

export default class GameMap extends Component {
    
    GameBoard() {
        const handleCountryClick = (e) => {
            const countryId = e.target.id;
            if (countryId) {
            console.log(`Pays cliqué : ${countryId}`);
            // Ici, tu peux récupérer les infos du pays via ton API Go
            }
        };

        return (
            <div>
            <EuropeMap onClick={handleCountryClick} style={{ width: '90%', height: 'auto' }} />
            </div>
        );
    }

    render() {
        return (
        <div>
        <h1>SVG Test</h1>
        {this.GameBoard()}
        </div>
    );
    }
}