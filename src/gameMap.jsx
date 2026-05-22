import React, {Component} from 'react';
import EuropeMap from '../assets/europeMap.svg';

export default class GameMap extends Component {

    state = { countryInfo: null, countryName: null, mapInfos: {}, gamePhase: '', currentUser: null, playerGold: 0};
    
    // la map apparait 
    componentDidMount() {
        this.fetchMapInfos();
        this.fetchCurrentUser();
        this.interval = setInterval(() => this.fetchMapInfos(), 5000);
    }

    // quand la map disparait
    componentWillUnmount() {
        clearInterval(this.interval);
    }

    fetchCurrentUser() {
        fetch('/me')
            .then(res => res.ok ? res.json() : null)
            .then(data => this.setState({ 
                currentUser: data?.username ?? null,
                playerGold: data?.gold ?? 0
            }))
            .catch(() => this.setState({ currentUser: null }));
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
                    if (this.state.countryName === country) {
                        this.setState({ countryInfo: null, countryName: null });
                        this.handleHighlightCountry(null);
                        return;
                    }

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
        const ressources = (info.produced_gold !== undefined && info.level !== undefined)
            ? `${(info.level+1) * info.produced_gold}K`
            : "Aucune ressource produite";

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
                <p style={{ margin: '0 0 6px 0' }}>
                    <strong>Production :</strong> {ressources}
                </p>
                <p style={{ margin:  0 }}>
                    <strong>Niveau :</strong> {info.level ?? 0} / 3
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
            .filter(([_, info]) => info.is_attacked || info.is_conquered)
            .map(([country, info]) => {
                const el = document.querySelector(`[data-country="${country}"]`);
                if (!el) return null;

                const rect = el.getBoundingClientRect();
                const containerRect = container.getBoundingClientRect();

                const emoji = info.is_conquered ? "🏛️" : "⚔️";

                const size = 40;

                const centerX = rect.left - containerRect.left + rect.width / 2;
                const centerY = rect.top - containerRect.top + rect.height / 2;

                return (
                    <div
                        key={country}
                        style={{
                            position: 'absolute',
                            left: centerX - size / 2,
                            top: centerY - size / 2,
                            fontSize: `${size}px`,
                            pointerEvents: 'none',
                            zIndex: 10,
                        }}
                    >
                        {emoji}
                    </div>
                );
            });
    }

    renderActionButton(leader, ressources) {
        const { gamePhase, currentUser, playerGold, countryInfo } = this.state;

        if (!currentUser || !countryInfo) return null;

        const level = countryInfo.level ?? 0;
        const upgradeCost =
            countryInfo.produced_gold * 1000 * 2 * (level + 1);

        const canUpgrade = level < 3 && playerGold >= upgradeCost;

        const button = (props, text) => (
            <button style={{ marginTop: '10px', ...props.style }} {...props}>
                {text}
            </button>
        );

        if (gamePhase === 'Attaque 🪖') {
            if (leader === "Non occupé" || leader === currentUser) return null;

            return countryInfo.attacked_by
                ? button({ disabled: true, style: { opacity: 0.6 } }, "Déjà en guerre")
                : button({ onClick: () => this.attackCountry() }, "Attaquer");
        }

        if (gamePhase !== 'Paix 🤝') return null;

        if (leader === "Non occupé") {
            return (countryInfo.attacked_by || countryInfo.conquered_by)
                ? button({ disabled: true, style: { opacity: 0.6 } }, "Déjà en conquête")
                : button(
                    { onClick: () => this.conquerCountry() },
                    `Conquérir (${ressources})`
                );
        }

        if (leader === currentUser) {
            return button(
                {
                    onClick: () => canUpgrade && this.upgradeCountry(),
                    disabled: !canUpgrade,
                    style: {
                        opacity: canUpgrade ? 1 : 0.5,
                        cursor: canUpgrade ? 'pointer' : 'not-allowed'
                    }
                },
                canUpgrade ? `Améliorer (${ressources})` : "Impossible"
            );
        }

        return null;
    }

    attackCountry() {
        const country = this.state.countryName;

        fetch(`/attackCountry?country=${country}`, {
            method: 'POST',
            credentials: 'include'
        })
            .then(async res => {
                const text = await res.text();
                if (!res.ok) throw new Error(text);
                return JSON.parse(text);
            })
            .then(() => {
                this.setState(prev => ({
                    countryInfo: {
                        ...prev.countryInfo,
                        attacked_by: prev.currentUser
                    }
                }));

                return fetch(`/getState?country=${country}`);
            })
            .then(res => res.json())
            .then(data => {
                this.setState({ countryInfo: data });
                this.fetchMapInfos();
            })
            .catch(err => {
                alert(err.message);
            });
    }

    conquerCountry() {
        const country = this.state.countryName;

        fetch(`/conquerCountry?country=${country}`, {
            method: 'POST',
            credentials: 'include'
        })
            .then(async res => {
                const text = await res.text();
                if (!res.ok) throw new Error(text);
                return JSON.parse(text);
            })
            .then(() => {
                this.setState(prev => ({
                    countryInfo: {
                        ...prev.countryInfo,
                        conquered_by: prev.currentUser
                    }
                }));

                return fetch(`/getState?country=${country}`);
            })
            .then(res => res.json())
            .then(data => {
                this.setState({ countryInfo: data });
                this.fetchMapInfos();
            })
            .catch(err => {
                alert(err.message);
            });
    }

    upgradeCountry() {
        const country = this.state.countryName;

        fetch(`/upgradeCountry?country=${country}`, {
            method: 'POST',
            credentials: 'include'
        })
            .then(async res => {
                const text = await res.text();

                if (!res.ok) {
                    throw new Error(text);
                }

                return JSON.parse(text);
            })
            .then(data => {
                console.log("Upgrade réussi :", data);

                // refresh infos du pays
                return fetch(`/getState?country=${country}`);
            })
            .then(res => res.json())
            .then(data => {
                this.setState({ countryInfo: data });
                this.fetchMapInfos();
            })
            .catch(err => {
                console.error("Erreur upgrade :", err);
                alert(err.message);
            });
    }

    render() {
        return (
            <div style={{ overflowX: 'hidden', margin: 0 }}>
                    {this.GameBoard()}
            </div>
        );
    }
}