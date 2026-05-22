import React, { Component } from 'react';
import { AuthContext } from './authContext';

export default class Login extends Component {
    static contextType = AuthContext;

    render() {
        const isLogged = Boolean(this.context.currentUser);

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