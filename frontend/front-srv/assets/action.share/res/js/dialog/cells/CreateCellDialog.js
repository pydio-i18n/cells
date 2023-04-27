/*
 * Copyright 2007-2017 Charles du Jeu - Abstrium SAS <team (at) pyd.io>
 * This file is part of Pydio.
 *
 * Pydio is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */

import React from 'react';
import PropTypes from 'prop-types';

import Pydio from 'pydio'
import {FontIcon, FlatButton, RaisedButton} from 'material-ui'
import {muiThemeable} from 'material-ui/styles'
import SharedUsers from './SharedUsers'
import NodesPicker from './NodesPicker'
import CellModel from 'pydio/model/cell'
import CellBaseFields from "./CellBaseFields";

/**
 * Dialog for letting users create a workspace
 */
class CreateCellDialog extends React.Component {
    static childContextTypes = {
        messages:PropTypes.object,
        getMessage:PropTypes.func,
        isReadonly:PropTypes.func
    };

    state = {step:'users', model:new CellModel(), saving: false};

    getChildContext() {
        const messages = this.props.pydio.MessageHash;
        return {
            messages: messages,
            getMessage: function(messageId, namespace='share_center'){
                try{
                    return messages[namespace + (namespace?".":"") + messageId] || messageId;
                }catch(e){
                    return messageId;
                }
            },
            isReadonly: function(){
                return false;
            }.bind(this)
        };
    }

    componentDidMount() {
        //this.refs.title.focus();
        this.state.model.observe('update', ()=>{this.forceUpdate()});
    }

    componentWillUnmount() {
        this.state.model.stopObserving('update');
    }

    submit = () => {
        const {model} = this.state;
        this.setState({saving: true});
        model.save().then(result => {
            this.props.onDismiss();
            this.setState({saving: false});
        }).catch(reason => {
            pydio.UI.displayMessage('ERROR', reason.message);
            this.setState({saving: false});
        });
    };

    m = (id) => {
        return this.props.pydio.MessageHash['share_center.' + id];
    };

    computeSummaryString = () => {
        const {model} = this.state;
        let users = 0;
        let groups = 0;
        let teams = 0;
        let userString = [];
        const objs = model.getAcls();
        Object.keys(objs).map(k => {
            const acl = objs[k];
            if(acl.Group) groups ++;
            else if(acl.Role) teams ++;
            else users ++;
        });
        if(users) userString.push( users + ' ' + this.m(270));
        if(groups) userString.push( groups + ' ' + this.m(271));
        if(teams) userString.push( teams + ' ' + this.m(272));
        let finalString;
        if (userString.length === 3) {
            finalString = userString[0] + ', ' + userString[1] + this.m(274) + userString[3];
        } else if (userString.length === 0) {
            finalString = this.m(273);
        } else {
            finalString = userString.join(this.m(274));
        }
        return this.m(269).replace('%USERS', finalString);
    };

    render() {

        let buttons = [];
        let content;
        let hPadding = '20px'
        const {pydio, muiTheme} = this.props;
        const {step, model, saving} = this.state;
        let dialogLabel = pydio.MessageHash['418'];
        if(step !== 'users'){
            dialogLabel = model.getLabel();
        }


        if (step === 'users'){

            content = (
                <div>
                    <div>{this.m(275)}</div>
                    <CellBaseFields
                        pydio={pydio}
                        model={model}
                        style={{padding: 0}}
                        muiTheme={muiTheme}
                        labelFocus={true}
                        labelEnter={() => this.submit()}
                        createLabels={true}
                    />
                </div>
            );

            if(model.getLabel()){
                buttons.push(<FlatButton
                    key="quick"
                    primary={true}
                    disabled={!model.getLabel() || saving}
                    label={this.m('cells.create.advanced')} // Advanced
                    onClick={()=>{this.setState({step:'data'})}} />
                );
                buttons.push(<span style={{display:'inline-block', margin: '0  10px', fontSize: 14, fontWeight: 500, color: '#9E9E9E'}}>{this.m('cells.create.buttons.separator')}</span>);
            }

            buttons.push(<RaisedButton
                key="next1"
                disabled={!model.getLabel() || saving}
                primary={true}
                label={this.m(279)} // Create Cell
                onClick={()=>{this.submit()}} />
            );


        } else if(step === 'data') {

            content = (
                <div>
                    <h5 style={{marginTop: -10, padding:'0 10px'}}>{this.m(278)}</h5>
                    <SharedUsers
                        pydio={pydio}
                        cellAcls={model.getAcls()}

                        excludes={[pydio.user.id]}
                        onUserObjectAdd={model.addUser.bind(model)}
                        onUserObjectRemove={model.removeUser.bind(model)}
                        onUserObjectUpdateRight={model.updateUserRight.bind(model)}
                    />
                </div>
            );
            hPadding = '10px'

            buttons.push(<FlatButton key="prev1" primary={false} label={pydio.MessageHash['304']} onClick={()=>{this.setState({step:'users'})}} />);
            buttons.push(<FlatButton key="next2" primary={true} label={pydio.MessageHash['179']} onClick={()=>this.setState({step:'label'})} />);

        } else {

            content = (
                <div>
                    <h5 style={{marginTop: -10}}>{this.m('cells.create.title.fill.folders')}</h5>
                    <div style={{color: 'var(--md-sys-color-outline)', paddingTop: 10}}>{this.computeSummaryString()}</div>
                    <div style={{paddingTop: 16}}>
                        <NodesPicker pydio={pydio} model={model}/>
                    </div>
                </div>
            );

            buttons.push(<FlatButton key="prev2" primary={false} label={pydio.MessageHash['304']} onClick={()=>{this.setState({step:'data'})}} />);
            buttons.push(<RaisedButton key="submit" disabled={saving} primary={true} label={this.m(279)} onClick={this.submit.bind(this)} />);

        }

        return (
            <div style={{width: 380, fontSize: 13, display:'flex', flexDirection:'column', minHeight: 300}}>
                <div style={{display:'flex', alignItems:'center', paddingLeft: 20}}>
                    <FontIcon className={"icomoon-cells-full-plus"}/>
                    <div style={{padding: 20, fontSize: 22}}>{dialogLabel}</div>
                </div>
                <div style={{padding: '20px '+hPadding+' 10px', flex:1}}>
                    {content}
                </div>
                <div style={{padding:'12px 16px', display: 'flex', alignItems: 'center', justifyContent: 'flex-end'}}>
                    {buttons}
                </div>
            </div>
        );

    }
}

CreateCellDialog = muiThemeable()(CreateCellDialog);
export {CreateCellDialog as default}