import React from 'react'
import PropTypes from 'prop-types';
import Pydio from 'pydio'
import User from '../model/User'
import {IconMenu, IconButton, MenuItem, FlatButton} from 'material-ui';
const {FormPanel} = Pydio.requireLib('form');
import {profileToLabel} from "../../board/Dashboard";

class UserInfo extends React.Component {

    constructor(props){
        super(props);
        this.state = {
            parameters: []
        };
        AdminComponents.PluginsLoader.getInstance(props.pydio).formParameters('//global_param[contains(@scope,"user")]|//param[contains(@scope,"user")]').then(params => {
            this.setState({parameters: params});
        })

    }

    getBinaryContext(){
        const {user} = this.props;
        return "user_id="+user.getIdmUser().Login + (user.getIdmUser().Attributes && user.getIdmUser().Attributes['avatar'] ? '?'+user.getIdmUser().Attributes['avatar'] : '');
    }

    getPydioRoleMessage(messageId){
        const {pydio} = this.props;
        return pydio.MessageHash['role_editor.' + messageId] || messageId;
    }

    onParameterChange(paramName, newValue, oldValue){
        const {user} = this.props;
        const {parameters} = this.state;
        const params = parameters.filter(p => p.name === paramName);
        const idmUser = user.getIdmUser();
        const role = user.getRole();
        // do something
        if(paramName === 'displayName' || paramName === 'email' || paramName === 'profile' || paramName === 'avatar'){
            idmUser.Attributes[paramName] = newValue;
        } else if (params.length && params[0].aclKey) {
            role.setParameter(params[0].aclKey, newValue);
        }
    }

    buttonCallback(action){
        const {user} = this.props;
        if(action === "update_user_pwd"){
            this.props.pydio.UI.openComponentInModal('AdminPeople', 'Editor.User.UserPasswordDialog', {user: user});
        }else{
            const idmUser = user.getIdmUser();
            const lockName = action === 'user_set_lock-lock' ? 'logout' : 'pass_change';
            let currentLocks = [];
            if(idmUser.Attributes['locks']){
                const test = JSON.parse(idmUser.Attributes['locks']);
                if(test && typeof test === "object"){
                    currentLocks = test;
                }
            }
            if(currentLocks.indexOf(lockName) > - 1){
                currentLocks = currentLocks.filter(l => l !== lockName);
                if(action === 'user_set_lock-lock'){
                    // Reset also the failedConnections attempts
                    delete idmUser.Attributes["failedConnections"];
                }
            } else {
                currentLocks.push(lockName);
            }
            idmUser.Attributes['locks'] = JSON.stringify(currentLocks);
            user.save();
        }
    }

    render(){

        const {user, pydio, adminStyles} = this.props;
        const {parameters} = this.state;
        if(!parameters){
            return <div>Loading...</div>;
        }

        let values = {profiles:[]};
        let locks = [];
        let hidden;

        if(user){
            // Compute values
            const idmUser = user.getIdmUser();
            const role = user.getRole();
            hidden = idmUser.Attributes['hidden'] === 'true'

            if(idmUser.Attributes['locks']){
                locks = JSON.parse(idmUser.Attributes['locks']) || [];
                if (typeof locks === 'object' && locks.length === undefined){ // Backward compat issue
                    let arrL = [];
                    Object.keys(locks).forEach(k => {
                        if(locks[k] === true) {
                            arrL.push(k);
                        }
                    });
                    locks = arrL;
                }
            }

            const attributes = idmUser.Attributes || {};
            values = {
                ...values,
                avatar: attributes['avatar'],
                displayName: attributes['displayName'],
                email: attributes['email'],
                profile: attributes['profile'],
                login: idmUser.Login
            };
            parameters.map(p => {
                if(p.aclKey && role.getParameterValue(p.aclKey)){
                    values[p.name] = role.getParameterValue(p.aclKey);
                }
            });
        }
        const profileChoices = ['admin', 'standard', 'shared'].map(p => p+'|'+profileToLabel(p, (i)=>pydio.MessageHash['settings.' + i]||i )).join(',')
        let params = [
            {name:"login", label:this.getPydioRoleMessage('21'),description:pydio.MessageHash['pydio_role.31'],"type":"string", readonly:true},
            {name:"profile", label:this.getPydioRoleMessage('22'), description:pydio.MessageHash['pydio_role.32'],"type":"select", choices:profileChoices},
            ...parameters
        ];
        if(hidden) {
            const allowed = ['login', 'avatar']
            params = params.filter(p => allowed.indexOf(p.name) > -1)
        }

        const secuActionsDisabled = (user.getIdmUser().Login === pydio.user.id)
        const buttons = [
            {label:'25', callback:'update_user_pwd'},
            {label:locks.indexOf('logout') > -1?'27':'26', callback:'user_set_lock-lock', active: locks.indexOf('logout') > -1},
            {label:locks.indexOf('pass_change') > -1?'28b':'28', callback:'user_set_lock-pass_change', active: locks.indexOf('pass_change') > -1}
        ]

        return (
            <div>
                <h3 className={"paper-right-title"}>
                    {pydio.MessageHash['pydio_role.24']}
                    <div className={"section-legend"}>{pydio.MessageHash['pydio_role.54']}</div>
                </h3>
                {!secuActionsDisabled &&
                    <div style={{padding: '10px 16px 0'}}>
                        {buttons.map(b => {
                            let ss = b.active? {backgroundColor: '#e53935'}: {}
                            return (
                                <FlatButton
                                    disabled={secuActionsDisabled}
                                    label={this.getPydioRoleMessage(b.label)}
                                    onClick={() => this.buttonCallback(b.callback)}
                                    {...adminStyles.props.header.flatButton}
                                    {...ss}
                                />
                            );
                        })}
                    </div>
                }
                <div className={"paper-right-block"} style={{padding: '6px 6px 2px'}}>
                    <FormPanel
                        parameters={params}
                        onParameterChange={this.onParameterChange.bind(this)}
                        values={values}
                        depth={-2}
                        variant={'v2'}
                        variantShowLegend={true}
                        binary_context={this.getBinaryContext()}
                    />
                </div>
            </div>
        );


    }

}

UserInfo.PropTypes = {
    pydio: PropTypes.instanceOf(Pydio).isRequired,
    pluginsRegistry: PropTypes.instanceOf(XMLDocument),
    user: PropTypes.instanceOf(User),
};

export {UserInfo as default}