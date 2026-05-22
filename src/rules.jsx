import React, { Component } from 'react';

export default class Rules extends Component {
    render() {
        return (
            <div style={{
                position: 'absolute',
                top: 10,
                right: 300,
                zIndex: 99999
            }}>
                <a
                    href="/rules.html"
                    style={{
                        display: 'inline-block',
                        padding: '18px 26px',
                        fontSize: '22px',
                        fontWeight: 'bold',
                        border: '2px solid black',
                        backgroundColor: '#e6e6e6', // même fond que login/logout
                        textDecoration: 'none',
                        color: 'black',
                        cursor: 'pointer'
                    }}
                >
                    Règles du jeu
                </a>
            </div>
        );
    }
}