import React, {Component} from 'react';
import { AuthContext } from './authContext';
import EuropeMap from '../assets/europeMap.svg';
import DeployForm from './deployForm';

export default class GameMap extends Component {

    static contextType = AuthContext;

    state = { countryInfo: null, countryName: null, mapInfos: {}, gamePhase: '', deployView: false };

    // la map apparait 
    componentDidMount() {
        this.fetchMapInfos();
        this.interval = setInterval(() => this.fetchMapInfos(), 5000);
        this.popupInterval = setInterval(() => this.fetchPhasePopup(), 6000);
    }

    // quand la map disparait
    componentWillUnmount() {
        clearInterval(this.interval);
        clearInterval(this.popupInterval);
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

    // Récupère le popup de fin de phase et l'affiche si nécessaire
    fetchPhasePopup() {
        fetch('/getPhasePopup')
            .then(res => res.json())
            .then(data => {
                if (!data) return;

                const hasAny = (arr) => Array.isArray(arr) && arr.length > 0;

                const shouldShow =
                    hasAny(data.conquered) ||
                    hasAny(data.gained_attack) ||
                    hasAny(data.lost_attack) ||
                    hasAny(data.lost_to_weather) ||
                    hasAny(data.improved);

                if (shouldShow) {
                    this.setState({ phasePopup: data });
                } else {
                    this.setState({ phasePopup: null }); // IMPORTANT
                }
            })
            .catch(err => console.error('Erreur fetchPhasePopup:', err));
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
        if (!this.state.countryInfo) return null;
        const { countryInfo, countryName, deployView } = this.state;

        const leader = countryInfo.leader_id ?? "Non occupé";
        const meteo = countryInfo.meteo ? `${countryInfo.meteo.temperature}°C - ${countryInfo.meteo.condition}` : "Aucune donnée météo";
        const ressources = (countryInfo.produced_gold !== undefined && countryInfo.level !== undefined)
            ? `${(countryInfo.level + 1) * countryInfo.produced_gold}K`
            : "Aucune ressource produite";

        return (
            <div style={{
                position: 'fixed', bottom: '20px', left: '20px',
                padding: '12px 16px', border: '1px solid black',
                backgroundColor: 'white', borderRadius: '8px',
                zIndex: 1000, minWidth: '200px',
            }}>
                {this.renderCloseButton()}
                {deployView ? this.renderDeployView() : (
                    <>
                        <p style={{ margin: '0 0 6px 0' }}><strong>Pays :</strong> {countryName}</p>
                        <p style={{ margin: '0 0 6px 0' }}><strong>Leader :</strong> {leader}</p>
                        <p style={{ margin: '0 0 6px 0' }}><strong>Météo :</strong> {meteo}</p>
                        <p style={{ margin: '0 0 6px 0' }}><strong>Production :</strong> {ressources}</p>
                        <p style={{ margin: 0 }}><strong>Niveau :</strong> {countryInfo.level ?? 0} / 3</p>
                        {this.renderActionButton(leader, ressources)}
                        {this.renderDeployButton()}
                    </>
                )}
            </div>
        );
    }

    renderDeployButton() {
        const { countryInfo, gamePhase } = this.state;
        const { currentUser } = this.context;
        if (!currentUser || !countryInfo) return null;
        if (gamePhase !== 'Attaque 🪖') return null;

        const isLeader = countryInfo.leader_id === currentUser;
        const isAttacker = countryInfo.attacked_by === currentUser;

        if (!isLeader && !isAttacker) return null;

        const mode = isLeader ? 'defending' : 'attacking';
        const slot = isLeader ? countryInfo.troops_defending : countryInfo.troops_attacking;
        const alreadyDeployed = slot?.count > 0;

        return (
            <button
                style={{ marginTop: '8px', display: 'block', opacity: alreadyDeployed ? 0.5 : 1, cursor: alreadyDeployed ? 'not-allowed' : 'pointer' }}
                disabled={alreadyDeployed}
                onClick={() => !alreadyDeployed && this.setState({ deployView: { mode } })}
            >
                {alreadyDeployed ? "Déploiement déjà fait" : (isLeader ? "Déployer la Défense" : "Déployer l'Attaque")}
            </button>
        );
    }

    renderDeployView() {
        const { deployView } = this.state;
        const { mode } = deployView;

        return (
            <DeployForm
                mode={mode}
                countryName={this.state.countryName}
                onBack={() => this.setState({ deployView: false })}
                onDeployed={() => {
                    fetch(`/getState?country=${this.state.countryName}`)
                        .then(res => res.json())
                        .then(data => this.setState({ countryInfo: data, deployView: false }));
                }}
            />
        );
    }

    closeInfo() {
        this.setState({ countryInfo: null, countryName: null, deployView: false });
        this.handleHighlightCountry(null);
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
        const { gamePhase, countryInfo, mapInfos, countryName } = this.state;
        const { currentUser, playerGold } = this.context;

        if (!currentUser || !countryInfo) return null;

        const level = countryInfo.level ?? 0;
        const production = (level + 1) * countryInfo.produced_gold;
        const upgradeCost = countryInfo.produced_gold * 1000 * 2 * (level + 1);

        const isMaxLevel = level >= 3;
        const hasEnoughGold = playerGold >= upgradeCost;

        const alreadyAttacking = Object.entries(mapInfos).find(
            ([name, info]) => info.attacked_by === currentUser && name !== countryName
        )?.[0] ?? null;

        const alreadyConquering = Object.entries(mapInfos).find(
            ([name, info]) => info.conquered_by === currentUser && name !== countryName
        )?.[0] ?? null;

        const button = (props, text) => (
            <button style={{ marginTop: '10px', ...props.style }} {...props}>
                {text}
            </button>
        );

        if (gamePhase === 'Attaque 🪖') {
            if (leader === "Non occupé" || leader === currentUser) return null;

            if (countryInfo.attacked_by)
                return button({ disabled: true, style: { opacity: 0.6 } }, "Déjà en guerre");

            if (alreadyAttacking)
                return button({ disabled: true, style: { opacity: 0.6 } }, `Attaque en cours : ${alreadyAttacking}`);

            return button({ onClick: () => this.attackCountry() }, "Attaquer");
        }

        if (gamePhase !== 'Paix 🤝') return null;

        if (leader === "Non occupé") {
            if (countryInfo.attacked_by || countryInfo.conquered_by)
                return button({ disabled: true, style: { opacity: 0.6 } }, "Déjà en conquête");

            if (alreadyConquering)
                return button({ disabled: true, style: { opacity: 0.6 } }, `Conquête en cours : ${alreadyConquering}`);

            return button({ onClick: () => this.conquerCountry() }, "Conquérir");
        }

        if (leader === currentUser) {
            let label = "Améliorer";

            if (isMaxLevel) {
                label = "Déjà Niveau Max";
            } else if (!hasEnoughGold) {
                label = "Or insuffisant";
            } else {
                label = `Améliorer (${production}K) - Coût: ${Math.round((countryInfo.produced_gold * 1000 * 2 * (level + 1)) / 1000)}K`;
            }

            return button(
                {
                    onClick: () => {
                        if (!isMaxLevel && hasEnoughGold) this.upgradeCountry();
                    },
                    disabled: isMaxLevel || !hasEnoughGold,
                    style: {
                        opacity: isMaxLevel || !hasEnoughGold ? 0.5 : 1,
                        cursor: isMaxLevel || !hasEnoughGold ? 'not-allowed' : 'pointer'
                    }
                },
                label
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
                    {this.renderPhasePopup()}
            </div>
        );
    }

    // Affiche le popup de fin de phase
    renderPhasePopup() {
        const popup = this.state.phasePopup;
        if (!popup) return null;

        const section = (title, arr, renderItem) => {
            if (!Array.isArray(arr) || arr.length === 0) return null;

            return (
                <div style={{ marginBottom: '10px' }}>
                    <strong>{title}:</strong>
                    <div style={{ marginLeft: '10px' }}>
                        {arr.map((item, i) => (
                            <div key={i}>
                                {renderItem(item)}
                            </div>
                        ))}
                    </div>
                </div>
            );
        };

        const close = () => {
            fetch('/ackPhasePopup', { method: 'POST', credentials: 'include' })
                .catch(() => {});
            this.setState({ phasePopup: null });
            this.fetchMapInfos();
        };

        return (
            <div style={{ position: 'fixed', left: '50%', top: '20%', transform: 'translateX(-50%)', zIndex: 2000 }}>
                <div style={{ background: 'white', border: '2px solid black', padding: '16px', minWidth: '420px', borderRadius: '8px' }}>

                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <h3 style={{ margin: 0 }}>Rapport de fin de phase</h3>
                        <button onClick={close} style={{ cursor: 'pointer' }}>✕</button>
                    </div>

                    <div style={{ marginTop: '10px' }}>

                        {section('Pays conquis', popup.conquered, (e) => (
                            <span>{e.country} → {e.by}</span>
                        ))}

                        {section('Attaques réussies', popup.gained_attack, (e) => (
                            <span>{e.attacker} a pris {e.country} (ancien: {e.previous_owner})</span>
                        ))}

                        {section('Attaques ratées', popup.lost_attack, (e) => (
                            <span>{e.attacker} a échoué contre {e.country} (owner: {e.owner})</span>
                        ))}

                        {section('Météo destructrice', popup.lost_to_weather, (e) => (
                            <span>{e.country} perdu par {e.previous_owner} ({e.reason})</span>
                        ))}

                        {section('Améliorations', popup.improved, (e) => (
                            <span>{e.country} ({e.owner}) : lvl {e.old_level} → {e.new_level}</span>
                        ))}

                        {!(
                            (popup.conquered?.length ?? 0) ||
                            (popup.gained_attack?.length ?? 0) ||
                            (popup.lost_attack?.length ?? 0) ||
                            (popup.lost_to_weather?.length ?? 0) ||
                            (popup.improved?.length ?? 0)
                        ) && (
                            <div>Aucun changement.</div>
                        )}

                    </div>
                </div>
            </div>
        );
    }
}