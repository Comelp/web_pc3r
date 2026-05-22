import React, { Component } from 'react';

export default class Login extends Component {

    state = {
        logged: false
    };

    componentDidMount() {
        fetch('/me', { credentials: 'include' })
            .then(res => {
                if (!res.ok) throw new Error();
                return res.json();
            })
            .then(() => this.setState({ logged: true }))
            .catch(() => this.setState({ logged: false }));
    }

    render() {
        const isLogged = this.state.logged;

        return (
            <div style={{
                position: 'absolute',
                top: 10,
                right: 10,
                zIndex: 99999
            }}>
                <a
                    href={isLogged ? "/logout" : "/login.html"}
                    style={{
                        display: 'inline-block',
                        padding: '18px 26px',
                        fontSize: '22px',
                        fontWeight: 'bold',
                        border: '2px solid black',
                        backgroundColor: '#e6e6e6',
                        textDecoration: 'none',
                        color: isLogged ? 'red' : '#1a73e8',
                        cursor: 'pointer'
                    }}
                >
                    {isLogged ? "Se déconnecter" : "Se connecter"}
                </a>
            </div>
        );
    }
}