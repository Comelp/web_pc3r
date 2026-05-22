import React, {Component} from 'react';
import EuropeMap from '../assets/europeMap.svg';

export default class GameMap extends Component {

    state = { countryInfo: null, countryName: null };
    
    GameBoard() {
        const handleCountryClick = (e) => {
            const el = e.nativeEvent.target.closest('[data-country]');

            if (!el) return;

            const country = el.dataset.country;

            console.log("demande d'informations pour :'"+country+"'");

            fetch(`/getState?country=${country}`)
                .then(res => {
                    if (!res.ok) {
                        throw new Error("Erreur serveur: " + res.status);
                    }
                    return res.json();
                })
                .then(data => {
                    this.setState({ countryInfo: data, countryName: country });
                    this.handleHighlightCountry(country);
                })
                .catch(err => console.error("Erreur fetch :", err));
        };

        return (
            <div>
                <div style={{ border: '3px solid black', display: 'inline-block', width: '99%' }}>
                    <EuropeMap
                        onClick={handleCountryClick}
                        style={{ width: '98vw', height: 'auto', display: 'block' }}
                    />
                </div>
                {this.renderInfosCountry()}
            </div>
        );
    }

    closeInfo() {
        this.setState({ countryInfo: null, countryName: null });
        this.handleHighlightCountry(null);
    }

    renderCloseButton() {
        return (
            <span
                onClick={() => this.closeInfo()}
                style={{
                    position: 'absolute',
                    top: '6px',
                    right: '8px',
                    cursor: 'pointer',
                    color: 'red',
                    fontWeight: 'bold',
                    fontSize: '14px',
                    lineHeight: 1,
                }}
            >✕</span>
        );
    }

    handleHighlightCountry(country) {
        const prev = document.querySelector('[data-country].selected');
        if (prev) {
            prev.style.fill = '#c0c0c0';
            prev.classList.remove('selected');
        }

        if (country) {
            const el = document.querySelector(`[data-country="${country}"]`);
            if (el) {
                el.style.fill = 'rgba(180, 73, 192, 0.5)';
                el.classList.add('selected');
            }
        }
    }

    renderInfosCountry() {
        if (!this.state.countryInfo) {
            return null;
        }
        const info = this.state.countryInfo;

        const leader = info.leader_id ? info.leader_id : "Non occupé";
        const meteo = info.meteo ? (`${info.meteo.temperature}°C - ${info.meteo.condition}`) : "Aucune donnée météo";

        return (
            <div style={{
                position: 'fixed',
                bottom: '20px',
                left: '20px',
                padding: '12px 16px',
                border: '1px solid black',
                backgroundColor: 'white',
                borderRadius: '8px',
                zIndex: 1000,
                minWidth: '200px',
            }}>
                {this.renderCloseButton()}
                <p style={{ margin: '0 0 6px 0' }}>
                    <strong>Pays :</strong> {this.state.countryName}
                </p>
                <p style={{ margin: '0 0 6px 0' }}>
                    <strong>Leader :</strong> {leader}
                </p>
                <p style={{ margin: 0 }}>
                    <strong>Météo :</strong> {meteo}
                </p>
            </div>
        );
    }

    render() {
        return (
            <div style={{ overflowX: 'hidden', margin: 0 }}>
                <h1>Carte</h1>
                    {this.GameBoard()}
            </div>
        );
    }
}