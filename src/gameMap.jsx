import React, {Component} from 'react';
import EuropeMap from '../assets/europeMap.svg';

export default class GameMap extends Component {

    state = { countryInfo: null, countryName: null, mapInfos: {}, gamePhase: ''};
    
    // la map apparait 
    componentDidMount() {
        this.fetchMapInfos();
        this.interval = setInterval(() => this.fetchMapInfos(), 5000);
    }

    // quand la map disparait
    componentWillUnmount() {
        clearInterval(this.interval);
    }


    fetchMapInfos() {
        fetch('/getMapInfos')
            .then(res => res.json())
            .then(data => {
                this.setState({ mapInfos: data }, () => this.applyCountriesColor());
                return fetch('/getPhase');
            })
            .then(res => res.text())
            .then(text => {
                try {
                    const data = JSON.parse(text);
                    this.setState({ gamePhase: data.phase });
                } catch (err) {
                    console.error('Invalid JSON from /getPhase:', text, err);
                }
            })
            .catch(err => console.error("Erreur fetchMapInfos :", err));
    }

    applyCountriesColor() {
        const { mapInfos } = this.state;
        Object.entries(mapInfos).forEach(([country, info]) => {
            const el = document.querySelector(`[data-country="${country}"]`);
            if (!el) return;
            if (el.classList.contains('selected')) return; // ne pas écraser la sélection en cours
            el.style.fill = info.color ?? '#c0c0c0';     // couleur de base sinon
        });
    }

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
                <div
                    className="map-container"
                    style={{ position: 'relative', border: '3px solid black', display: 'inline-block', width: '99%' }}
                >
                    <EuropeMap
                        onClick={handleCountryClick}
                        style={{ width: '98vw', height: 'auto', display: 'block' }}
                    />
                    {this.renderCombatLogos()}
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
            const prevName = prev.dataset.country;
            const prevInfo = this.state.mapInfos[prevName];
            prev.style.fill = prevInfo?.color ?? '#c0c0c0';
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
        const ressources = info.produced_gold ? (`${info.produced_gold}K`) : "Aucune ressource produite";


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
                <p style={{ margin: '0 0 6px 0' }}>
                    <strong>Météo :</strong> {meteo}
                </p>
                <p style={{ margin: 0 }}>
                    <strong>Production :</strong> {ressources}
                </p>
                {this.renderActionButton(leader, ressources)}
            </div>
        );
    }

    renderCombatLogos() {
        const { mapInfos } = this.state;
        const container = document.querySelector('.map-container');
        if (!container) return null;

        return Object.entries(mapInfos)
            .filter(([_, info]) => info.is_attacked)
            .map(([country]) => {
                const el = document.querySelector(`[data-country="${country}"]`);
                if (!el) return null;

                const rect = el.getBoundingClientRect();
                const containerRect = container.getBoundingClientRect();
                return (
                    <div
                        key={country}
                        style={{
                            position: 'absolute',
                            left: rect.left - containerRect.left + 2 * rect.width / 5,
                            top: rect.top - containerRect.top + 2 * rect.height / 5,
                            pointerEvents: 'none',
                            fontSize: '40px',
                            zIndex: 10,
                        }}
                    >
                        ⚔️
                    </div>
                );
            });
    }

    renderActionButton(leader, ressources) {
        const phase = this.state.gamePhase;

        if (phase === 'Paix 🤝') {
            if (leader === "Non occupé") {
                return <button onClick={() => this.conquerCountry()} style={{ marginTop: '10px' }}>Conquérir ({ressources})</button>;
            }
            if (leader === "Vous") {
                return <button style={{ marginTop: '10px' }}>Améliorer ({ressources})</button>;
            }
            return null;
        }

        if (phase === 'Attaque 🪖' && leader !== "Non occupé" && leader !== "Vous") {
            return <button style={{ marginTop: '10px' }}>Attaquer</button>;
        }

        return null;
    }

    render() {
        return (
            <div style={{ overflowX: 'hidden', margin: 0 }}>
                    {this.GameBoard()}
            </div>
        );
    }
}