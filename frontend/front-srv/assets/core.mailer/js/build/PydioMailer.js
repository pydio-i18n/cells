(function(f){if(typeof exports==="object"&&typeof module!=="undefined"){module.exports=f()}else if(typeof define==="function"&&define.amd){define([],f)}else{var g;if(typeof window!=="undefined"){g=window}else if(typeof global!=="undefined"){g=global}else if(typeof self!=="undefined"){g=self}else{g=this}g.PydioMailer = f()}})(function(){var define,module,exports;return (function e(t,n,r){function s(o,u){if(!n[o]){if(!t[o]){var a=typeof require=="function"&&require;if(!u&&a)return a(o,!0);if(i)return i(o,!0);var f=new Error("Cannot find module '"+o+"'");throw f.code="MODULE_NOT_FOUND",f}var l=n[o]={exports:{}};t[o][0].call(l.exports,function(e){var n=t[o][1][e];return s(n?n:e)},l,l.exports,e,t,n,r)}return n[o].exports}var i=typeof require=="function"&&require;for(var o=0;o<r.length;o++)s(r[o]);return s})({1:[function(require,module,exports){
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

'use strict';

Object.defineProperty(exports, '__esModule', {
    value: true
});

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

var _createClass = (function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ('value' in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; })();

var _get = function get(_x10, _x11, _x12) { var _again = true; _function: while (_again) { var object = _x10, property = _x11, receiver = _x12; _again = false; if (object === null) object = Function.prototype; var desc = Object.getOwnPropertyDescriptor(object, property); if (desc === undefined) { var parent = Object.getPrototypeOf(object); if (parent === null) { return undefined; } else { _x10 = parent; _x11 = property; _x12 = receiver; _again = true; desc = parent = undefined; continue _function; } } else if ('value' in desc) { return desc.value; } else { var getter = desc.get; if (getter === undefined) { return undefined; } return getter.call(receiver); } } };

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError('Cannot call a class as a function'); } }

function _inherits(subClass, superClass) { if (typeof superClass !== 'function' && superClass !== null) { throw new TypeError('Super expression must either be null or a function, not ' + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _pydioHttpApi = require('pydio/http/api');

var _pydioHttpApi2 = _interopRequireDefault(_pydioHttpApi);

var _materialUi = require('material-ui');

var _pydioHttpRestApi = require('pydio/http/rest-api');

var styles = {
    chip: {
        marginRight: 4,
        marginBottom: 4
    },
    wrapper: {
        display: 'flex',
        flexWrap: 'wrap'
    },
    overlay: {
        position: 'absolute',
        top: 0,
        right: 0,
        left: 0,
        bottom: 0,
        backgroundColor: 'rgba(0, 0, 0, 0.33)',
        paddingTop: 77,
        zIndex: 100
    }
};

var DestBadge = (function (_React$Component) {
    _inherits(DestBadge, _React$Component);

    function DestBadge() {
        _classCallCheck(this, DestBadge);

        _get(Object.getPrototypeOf(DestBadge.prototype), 'constructor', this).apply(this, arguments);
    }

    _createClass(DestBadge, [{
        key: 'render',
        value: function render() {
            var userObject = this.props.user;
            return _react2['default'].createElement(
                'div',
                { className: "share-dialog user-badge user-type-" + (userObject.getTemporary() ? "tmp_user" : "user") },
                _react2['default'].createElement('span', { className: "avatar icon-" + (userObject.getTemporary() ? "envelope" : "user") }),
                _react2['default'].createElement(
                    'span',
                    { className: 'user-badge-label' },
                    userObject.getExtendedLabel() || userObject.getLabel()
                )
            );
        }
    }]);

    return DestBadge;
})(_react2['default'].Component);

var UserChip = (function (_React$Component2) {
    _inherits(UserChip, _React$Component2);

    function UserChip() {
        _classCallCheck(this, UserChip);

        _get(Object.getPrototypeOf(UserChip.prototype), 'constructor', this).apply(this, arguments);
    }

    _createClass(UserChip, [{
        key: 'render',
        value: function render() {
            var _props = this.props;
            var user = _props.user;
            var onRemove = _props.onRemove;

            var tmp = user.FreeValue;
            var label = undefined;
            if (tmp) {
                label = user.FreeValue;
            } else {
                if (user.Attributes && user.Attributes['displayName']) {
                    label = user.Attributes['displayName'];
                } else {
                    label = user.Login;
                }
            }

            var icon = _react2['default'].createElement(_materialUi.FontIcon, { className: "icon-" + (tmp ? "envelope" : "user") });
            var colors = _materialUi.Style.colors;

            return _react2['default'].createElement(
                _materialUi.Chip,
                {
                    backgroundColor: tmp ? colors.lightBlue100 : colors.blueGrey100,
                    onRequestDelete: function () {
                        onRemove();
                    },
                    style: styles.chip
                },
                _react2['default'].createElement(_materialUi.Avatar, { icon: icon, color: tmp ? 'white' : colors.blueGrey600, backgroundColor: tmp ? colors.lightBlue300 : colors.blueGrey300 }),
                label
            );
        }
    }]);

    return UserChip;
})(_react2['default'].Component);

var Email = (function () {
    function Email() {
        var subject = arguments.length <= 0 || arguments[0] === undefined ? null : arguments[0];
        var message = arguments.length <= 1 || arguments[1] === undefined ? null : arguments[1];
        var link = arguments.length <= 2 || arguments[2] === undefined ? null : arguments[2];

        _classCallCheck(this, Email);

        this._subjects = [];
        this._messages = [];
        this._targets = [];
        this._links = [];
        this.templateId = "";
        this.templateData = {};
        if (subject) {
            this._subjects.push(subject);
        }
        if (message) {
            this._messages.push(message);
        }
        if (link) {
            this._links.push(link);
        }
    }

    _createClass(Email, [{
        key: 'addTarget',
        value: function addTarget(userOrEmail) {
            var subject = arguments.length <= 1 || arguments[1] === undefined ? null : arguments[1];
            var message = arguments.length <= 2 || arguments[2] === undefined ? null : arguments[2];
            var link = arguments.length <= 3 || arguments[3] === undefined ? null : arguments[3];

            this._targets.push(userOrEmail);
            if (subject) {
                this._subjects.push(subject);
            }
            if (message) {
                this._messages.push(message);
            }
            if (link) {
                this._links.push(link);
            }
        }
    }, {
        key: 'setTemplateData',
        value: function setTemplateData(templateId, templateData) {
            this.templateId = templateId;
            this.templateData = templateData;
        }

        /**
         *
         * @param targets Array
         * @param subject string
         * @param templateId string
         * @param templateData Object
         * @return {Promise}
         */
    }, {
        key: 'postOne',
        value: function postOne(targets, subject, templateId, templateData) {
            var api = new _pydioHttpRestApi.MailerServiceApi(_pydioHttpApi2['default'].getRestClient());
            var email = new _pydioHttpRestApi.MailerMail();
            email.Subject = subject;
            email.To = targets.map(function (t) {
                return _pydioHttpRestApi.MailerUser.constructFromObject({ Uuid: t });
            });
            email.TemplateId = templateId;
            email.TemplateData = templateData;
            return api.send(email);
        }
    }, {
        key: 'post',
        value: function post() {
            var _this = this;

            var callback = arguments.length <= 0 || arguments[0] === undefined ? null : arguments[0];

            if (!this._subjects.length && !this.templateId || !this._targets.length || !this._messages.length) {
                throw new Error('Invalid data');
            }
            var templateData = _extends({}, this.templateData);
            var proms = [];

            if (this._messages.length > 1 && this._messages.length === this._targets.length && this._subjects.length === this._targets.length) {
                // Send as many emails as targets with their own messages
                this._targets.map(function (t, i) {
                    var subject = _this._subjects[i];
                    templateData = _extends({}, templateData, { Message: _this._messages[i] });
                    proms.push(_this.postOne([t], subject, _this.templateId, templateData));
                });
            } else {
                var subject = this._subjects[0];
                templateData['Message'] = this._messages[0];
                proms.push(this.postOne(this._targets, subject, this.templateId, templateData));
            }

            Promise.all(proms).then(function () {
                callback(true);
            })['catch'](function (e) {
                if (e.response && e.response.body && e.response.body.Title) {
                    e = new Error(e.response.body.Title);
                }
                callback(false, e);
            });
        }
    }]);

    return Email;
})();

var Pane = (function (_React$Component3) {
    _inherits(Pane, _React$Component3);

    function Pane(props) {
        _classCallCheck(this, Pane);

        if (props.showAddressBook === undefined) {
            props.showAddressBook = true;
        }
        _get(Object.getPrototypeOf(Pane.prototype), 'constructor', this).call(this, props);
        this.state = {
            users: this.props.users || {},
            subject: this.props.subject,
            message: this.props.message,
            errorMessage: null,
            posting: false
        };
    }

    _createClass(Pane, [{
        key: 'updateSubject',
        value: function updateSubject(event) {
            this.setState({ subject: event.currentTarget.value });
        }
    }, {
        key: 'updateMessage',
        value: function updateMessage(event) {
            this.setState({ message: event.currentTarget.value });
        }
    }, {
        key: 'addUser',
        value: function addUser(userObject) {
            var users = this.state.users;

            if (userObject.FreeValue) {
                users[userObject.FreeValue] = userObject;
            } else if (userObject.IdmUser) {
                users[userObject.IdmUser.Login] = userObject.IdmUser;
            }
            this.setState({ users: users, errorMessage: null });
        }
    }, {
        key: 'removeUser',
        value: function removeUser(userId) {
            var users = this.state.users;

            delete users[userId];
            this.setState({ users: users });
        }
    }, {
        key: 'getMessage',
        value: function getMessage(messageId) {
            var nameSpace = arguments.length <= 1 || arguments[1] === undefined ? undefined : arguments[1];

            try {
                if (nameSpace === undefined) {
                    nameSpace = 'core.mailer';
                }
                if (nameSpace) {
                    nameSpace += ".";
                }
                return pydio.MessageHash[nameSpace + messageId];
            } catch (e) {
                return messageId;
            }
        }
    }, {
        key: 'postEmail',
        value: function postEmail() {
            var _this2 = this;

            var repost = arguments.length <= 0 || arguments[0] === undefined ? false : arguments[0];
            var _state = this.state;
            var users = _state.users;
            var subject = _state.subject;
            var message = _state.message;

            if (!repost && this.refs.completer && this.refs.completer.getPendingSearchText && this.refs.completer.getPendingSearchText()) {
                this.refs.completer.onCompleterRequest(this.refs.completer.getPendingSearchText(), -1);
                setTimeout(function () {
                    return _this2.postEmail(true);
                }, 500);
                return;
            }
            if (!Object.keys(users).length) {
                this.setState({ errorMessage: this.getMessage(2) });
                return;
            }
            var _props2 = this.props;
            var link = _props2.link;
            var templateId = _props2.templateId;
            var templateData = _props2.templateData;

            var callback = function callback(res, err) {
                _this2.setState({ posting: false });
                if (res) {
                    _this2.props.pydio.UI.displayMessage('SUCCESS', _this2.props.pydio.MessageHash["core.mailer.1"].replace('%s', Object.keys(users).length));
                    if (_this2.props.onDismiss) {
                        _this2.props.onDismiss();
                    } else {
                        _this2.setState({ users: {}, subject: '', message: '' });
                    }
                } else {
                    _this2.props.pydio.UI.displayMessage('ERROR', _this2.props.pydio.MessageHash["core.mailer.sender.error"] + (err ? ' : ' + err.message : ''));
                }
            };
            this.setState({ posting: true });

            if (this.props.processPost) {
                this.props.processPost(Email, users, subject, message, link, callback);
                return;
            }

            var email = new Email(subject, message, link || null);
            Object.keys(users).forEach(function (k) {
                email.addTarget(k);
            });
            if (templateId) {
                email.setTemplateData(templateId, templateData);
            }
            email.post(callback);
        }
    }, {
        key: 'render',
        value: function render() {
            var _this4 = this;

            var _state2 = this.state;
            var users = _state2.users;
            var posting = _state2.posting;
            var errorMessage = _state2.errorMessage;
            var subject = _state2.subject;
            var message = _state2.message;

            var className = [this.props.className, "react-mailer", "reset-pydio-forms"].join(" ");
            var usersChips = Object.keys(users).map((function (uId) {
                var _this3 = this;

                return _react2['default'].createElement(UserChip, { key: uId, user: users[uId], onRemove: function () {
                        _this3.removeUser(uId);
                    } });
            }).bind(this));
            var errorDiv = undefined;
            if (errorMessage) {
                errorDiv = _react2['default'].createElement(
                    'div',
                    { style: { padding: '10px 20px', color: 'red' } },
                    errorMessage
                );
            }
            var style = _extends({
                margin: this.props.uniqueUserStyle ? 0 : 8
            }, this.props.style);
            var content = _react2['default'].createElement(
                _materialUi.Paper,
                { zDepth: this.props.zDepth !== undefined ? this.props.zDepth : 2, className: className, style: style },
                _react2['default'].createElement(
                    'h3',
                    { style: _extends({ padding: 20, color: 'rgba(0,0,0,0.87)', fontSize: 25, marginBottom: 0, paddingBottom: 10 }, this.props.titleStyle) },
                    this.props.panelTitle
                ),
                errorDiv,
                this.props.additionalPaneTop,
                !this.props.uniqueUserStyle && _react2['default'].createElement(
                    'div',
                    { className: 'users-block', style: _extends({ padding: '0 20px' }, this.props.usersBlockStyle) },
                    _react2['default'].createElement(PydioComponents.UsersCompleter, {
                        ref: 'completer',
                        fieldLabel: this.getMessage('8'),
                        usersOnly: true,
                        existingOnly: true,
                        freeValueAllowed: true,
                        onValueSelected: this.addUser.bind(this),
                        excludes: Object.keys(users),
                        renderSuggestion: function (userObject) {
                            return _react2['default'].createElement(DestBadge, { user: userObject });
                        },
                        pydio: pydio,
                        showAddressBook: this.props.showAddressBook,
                        underlineHide: true
                    }),
                    _react2['default'].createElement(
                        'div',
                        { style: styles.wrapper },
                        usersChips
                    )
                ),
                !this.props.uniqueUserStyle && _react2['default'].createElement(_materialUi.Divider, null),
                !this.props.templateId && _react2['default'].createElement(
                    'div',
                    { style: { padding: '0 20px' } },
                    _react2['default'].createElement(_materialUi.TextField, { fullWidth: true, underlineShow: false, floatingLabelText: this.getMessage('6'), value: subject, onChange: this.updateSubject.bind(this) })
                ),
                !this.props.templateId && _react2['default'].createElement(_materialUi.Divider, null),
                _react2['default'].createElement(
                    'div',
                    { style: _extends({ padding: '0 20px' }, this.props.messageBlockStyle) },
                    _react2['default'].createElement(_materialUi.TextField, {
                        fullWidth: true,
                        underlineShow: false,
                        floatingLabelText: this.getMessage('7'),
                        value: message,
                        multiLine: true,
                        onChange: this.updateMessage.bind(this),
                        rowsMax: 6
                    })
                ),
                this.props.additionalPaneBottom,
                _react2['default'].createElement(_materialUi.Divider, null),
                _react2['default'].createElement(
                    'div',
                    { style: { textAlign: 'right', padding: '8px 20px' } },
                    this.props.onDismiss && _react2['default'].createElement(_materialUi.FlatButton, { label: this.getMessage('54', ''), onTouchTap: this.props.onDismiss }),
                    !this.props.onDismiss && _react2['default'].createElement(_materialUi.FlatButton, { label: this.getMessage('216', ''), onTouchTap: function () {
                            _this4.setState({ users: {}, subject: '', message: '' });
                        } }),
                    _react2['default'].createElement(_materialUi.FlatButton, { disabled: posting, primary: true, label: this.getMessage('77', ''), onTouchTap: function (e) {
                            return _this4.postEmail();
                        } })
                )
            );
            if (this.props.overlay) {
                return _react2['default'].createElement(
                    'div',
                    { style: styles.overlay },
                    content
                );
            } else {
                return content;
            }
        }
    }]);

    return Pane;
})(_react2['default'].Component);

Pane.PropTypes = {
    message: _react2['default'].PropTypes.string,
    subject: _react2['default'].PropTypes.string,
    templateId: _react2['default'].PropTypes.string,
    templateData: _react2['default'].PropTypes.object,
    link: _react2['default'].PropTypes.string,
    onDismiss: _react2['default'].PropTypes.func,
    className: _react2['default'].PropTypes.string,
    overlay: _react2['default'].PropTypes.bool,
    uniqueUserStyle: _react2['default'].PropTypes.bool,
    users: _react2['default'].PropTypes.object,
    panelTitle: _react2['default'].PropTypes.string,
    zDepth: _react2['default'].PropTypes.number,
    showAddressBook: _react2['default'].PropTypes.bool,
    processPost: _react2['default'].PropTypes.func,
    additionalPaneTop: _react2['default'].PropTypes.instanceOf(_react2['default'].Component),
    additionalPaneBottom: _react2['default'].PropTypes.instanceOf(_react2['default'].Component)
};

var PreferencesPanel = (function (_React$Component4) {
    _inherits(PreferencesPanel, _React$Component4);

    function PreferencesPanel() {
        _classCallCheck(this, PreferencesPanel);

        _get(Object.getPrototypeOf(PreferencesPanel.prototype), 'constructor', this).apply(this, arguments);
    }

    _createClass(PreferencesPanel, [{
        key: 'render',
        value: function render() {
            return _react2['default'].createElement(
                'div',
                null,
                'Preferences Panel'
            );
        }
    }]);

    return PreferencesPanel;
})(_react2['default'].Component);

exports.Pane = Pane;
exports.PreferencesPanel = PreferencesPanel;

},{"material-ui":"material-ui","pydio/http/api":"pydio/http/api","pydio/http/rest-api":"pydio/http/rest-api","react":"react"}],2:[function(require,module,exports){
"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});

var _components = require("./components");

exports.Pane = _components.Pane;
exports.PreferencesPanel = _components.PreferencesPanel;

},{"./components":1}]},{},[2])(2)
});

//# sourceMappingURL=data:application/json;charset:utf-8;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIm5vZGVfbW9kdWxlcy9icm93c2VyLXBhY2svX3ByZWx1ZGUuanMiLCJqcy9idWlsZC9jb21wb25lbnRzLmpzIiwianMvYnVpbGQvaW5kZXguanMiXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IkFBQUE7QUNBQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTs7QUM1aEJBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0EiLCJmaWxlIjoiZ2VuZXJhdGVkLmpzIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXNDb250ZW50IjpbIihmdW5jdGlvbiBlKHQsbixyKXtmdW5jdGlvbiBzKG8sdSl7aWYoIW5bb10pe2lmKCF0W29dKXt2YXIgYT10eXBlb2YgcmVxdWlyZT09XCJmdW5jdGlvblwiJiZyZXF1aXJlO2lmKCF1JiZhKXJldHVybiBhKG8sITApO2lmKGkpcmV0dXJuIGkobywhMCk7dmFyIGY9bmV3IEVycm9yKFwiQ2Fubm90IGZpbmQgbW9kdWxlICdcIitvK1wiJ1wiKTt0aHJvdyBmLmNvZGU9XCJNT0RVTEVfTk9UX0ZPVU5EXCIsZn12YXIgbD1uW29dPXtleHBvcnRzOnt9fTt0W29dWzBdLmNhbGwobC5leHBvcnRzLGZ1bmN0aW9uKGUpe3ZhciBuPXRbb11bMV1bZV07cmV0dXJuIHMobj9uOmUpfSxsLGwuZXhwb3J0cyxlLHQsbixyKX1yZXR1cm4gbltvXS5leHBvcnRzfXZhciBpPXR5cGVvZiByZXF1aXJlPT1cImZ1bmN0aW9uXCImJnJlcXVpcmU7Zm9yKHZhciBvPTA7bzxyLmxlbmd0aDtvKyspcyhyW29dKTtyZXR1cm4gc30pIiwiLypcbiAqIENvcHlyaWdodCAyMDA3LTIwMTcgQ2hhcmxlcyBkdSBKZXUgLSBBYnN0cml1bSBTQVMgPHRlYW0gKGF0KSBweWQuaW8+XG4gKiBUaGlzIGZpbGUgaXMgcGFydCBvZiBQeWRpby5cbiAqXG4gKiBQeWRpbyBpcyBmcmVlIHNvZnR3YXJlOiB5b3UgY2FuIHJlZGlzdHJpYnV0ZSBpdCBhbmQvb3IgbW9kaWZ5XG4gKiBpdCB1bmRlciB0aGUgdGVybXMgb2YgdGhlIEdOVSBBZmZlcm8gR2VuZXJhbCBQdWJsaWMgTGljZW5zZSBhcyBwdWJsaXNoZWQgYnlcbiAqIHRoZSBGcmVlIFNvZnR3YXJlIEZvdW5kYXRpb24sIGVpdGhlciB2ZXJzaW9uIDMgb2YgdGhlIExpY2Vuc2UsIG9yXG4gKiAoYXQgeW91ciBvcHRpb24pIGFueSBsYXRlciB2ZXJzaW9uLlxuICpcbiAqIFB5ZGlvIGlzIGRpc3RyaWJ1dGVkIGluIHRoZSBob3BlIHRoYXQgaXQgd2lsbCBiZSB1c2VmdWwsXG4gKiBidXQgV0lUSE9VVCBBTlkgV0FSUkFOVFk7IHdpdGhvdXQgZXZlbiB0aGUgaW1wbGllZCB3YXJyYW50eSBvZlxuICogTUVSQ0hBTlRBQklMSVRZIG9yIEZJVE5FU1MgRk9SIEEgUEFSVElDVUxBUiBQVVJQT1NFLiAgU2VlIHRoZVxuICogR05VIEFmZmVybyBHZW5lcmFsIFB1YmxpYyBMaWNlbnNlIGZvciBtb3JlIGRldGFpbHMuXG4gKlxuICogWW91IHNob3VsZCBoYXZlIHJlY2VpdmVkIGEgY29weSBvZiB0aGUgR05VIEFmZmVybyBHZW5lcmFsIFB1YmxpYyBMaWNlbnNlXG4gKiBhbG9uZyB3aXRoIFB5ZGlvLiAgSWYgbm90LCBzZWUgPGh0dHA6Ly93d3cuZ251Lm9yZy9saWNlbnNlcy8+LlxuICpcbiAqIFRoZSBsYXRlc3QgY29kZSBjYW4gYmUgZm91bmQgYXQgPGh0dHBzOi8vcHlkaW8uY29tPi5cbiAqL1xuXG4ndXNlIHN0cmljdCc7XG5cbk9iamVjdC5kZWZpbmVQcm9wZXJ0eShleHBvcnRzLCAnX19lc01vZHVsZScsIHtcbiAgICB2YWx1ZTogdHJ1ZVxufSk7XG5cbnZhciBfZXh0ZW5kcyA9IE9iamVjdC5hc3NpZ24gfHwgZnVuY3Rpb24gKHRhcmdldCkgeyBmb3IgKHZhciBpID0gMTsgaSA8IGFyZ3VtZW50cy5sZW5ndGg7IGkrKykgeyB2YXIgc291cmNlID0gYXJndW1lbnRzW2ldOyBmb3IgKHZhciBrZXkgaW4gc291cmNlKSB7IGlmIChPYmplY3QucHJvdG90eXBlLmhhc093blByb3BlcnR5LmNhbGwoc291cmNlLCBrZXkpKSB7IHRhcmdldFtrZXldID0gc291cmNlW2tleV07IH0gfSB9IHJldHVybiB0YXJnZXQ7IH07XG5cbnZhciBfY3JlYXRlQ2xhc3MgPSAoZnVuY3Rpb24gKCkgeyBmdW5jdGlvbiBkZWZpbmVQcm9wZXJ0aWVzKHRhcmdldCwgcHJvcHMpIHsgZm9yICh2YXIgaSA9IDA7IGkgPCBwcm9wcy5sZW5ndGg7IGkrKykgeyB2YXIgZGVzY3JpcHRvciA9IHByb3BzW2ldOyBkZXNjcmlwdG9yLmVudW1lcmFibGUgPSBkZXNjcmlwdG9yLmVudW1lcmFibGUgfHwgZmFsc2U7IGRlc2NyaXB0b3IuY29uZmlndXJhYmxlID0gdHJ1ZTsgaWYgKCd2YWx1ZScgaW4gZGVzY3JpcHRvcikgZGVzY3JpcHRvci53cml0YWJsZSA9IHRydWU7IE9iamVjdC5kZWZpbmVQcm9wZXJ0eSh0YXJnZXQsIGRlc2NyaXB0b3Iua2V5LCBkZXNjcmlwdG9yKTsgfSB9IHJldHVybiBmdW5jdGlvbiAoQ29uc3RydWN0b3IsIHByb3RvUHJvcHMsIHN0YXRpY1Byb3BzKSB7IGlmIChwcm90b1Byb3BzKSBkZWZpbmVQcm9wZXJ0aWVzKENvbnN0cnVjdG9yLnByb3RvdHlwZSwgcHJvdG9Qcm9wcyk7IGlmIChzdGF0aWNQcm9wcykgZGVmaW5lUHJvcGVydGllcyhDb25zdHJ1Y3Rvciwgc3RhdGljUHJvcHMpOyByZXR1cm4gQ29uc3RydWN0b3I7IH07IH0pKCk7XG5cbnZhciBfZ2V0ID0gZnVuY3Rpb24gZ2V0KF94MTAsIF94MTEsIF94MTIpIHsgdmFyIF9hZ2FpbiA9IHRydWU7IF9mdW5jdGlvbjogd2hpbGUgKF9hZ2FpbikgeyB2YXIgb2JqZWN0ID0gX3gxMCwgcHJvcGVydHkgPSBfeDExLCByZWNlaXZlciA9IF94MTI7IF9hZ2FpbiA9IGZhbHNlOyBpZiAob2JqZWN0ID09PSBudWxsKSBvYmplY3QgPSBGdW5jdGlvbi5wcm90b3R5cGU7IHZhciBkZXNjID0gT2JqZWN0LmdldE93blByb3BlcnR5RGVzY3JpcHRvcihvYmplY3QsIHByb3BlcnR5KTsgaWYgKGRlc2MgPT09IHVuZGVmaW5lZCkgeyB2YXIgcGFyZW50ID0gT2JqZWN0LmdldFByb3RvdHlwZU9mKG9iamVjdCk7IGlmIChwYXJlbnQgPT09IG51bGwpIHsgcmV0dXJuIHVuZGVmaW5lZDsgfSBlbHNlIHsgX3gxMCA9IHBhcmVudDsgX3gxMSA9IHByb3BlcnR5OyBfeDEyID0gcmVjZWl2ZXI7IF9hZ2FpbiA9IHRydWU7IGRlc2MgPSBwYXJlbnQgPSB1bmRlZmluZWQ7IGNvbnRpbnVlIF9mdW5jdGlvbjsgfSB9IGVsc2UgaWYgKCd2YWx1ZScgaW4gZGVzYykgeyByZXR1cm4gZGVzYy52YWx1ZTsgfSBlbHNlIHsgdmFyIGdldHRlciA9IGRlc2MuZ2V0OyBpZiAoZ2V0dGVyID09PSB1bmRlZmluZWQpIHsgcmV0dXJuIHVuZGVmaW5lZDsgfSByZXR1cm4gZ2V0dGVyLmNhbGwocmVjZWl2ZXIpOyB9IH0gfTtcblxuZnVuY3Rpb24gX2ludGVyb3BSZXF1aXJlRGVmYXVsdChvYmopIHsgcmV0dXJuIG9iaiAmJiBvYmouX19lc01vZHVsZSA/IG9iaiA6IHsgJ2RlZmF1bHQnOiBvYmogfTsgfVxuXG5mdW5jdGlvbiBfY2xhc3NDYWxsQ2hlY2soaW5zdGFuY2UsIENvbnN0cnVjdG9yKSB7IGlmICghKGluc3RhbmNlIGluc3RhbmNlb2YgQ29uc3RydWN0b3IpKSB7IHRocm93IG5ldyBUeXBlRXJyb3IoJ0Nhbm5vdCBjYWxsIGEgY2xhc3MgYXMgYSBmdW5jdGlvbicpOyB9IH1cblxuZnVuY3Rpb24gX2luaGVyaXRzKHN1YkNsYXNzLCBzdXBlckNsYXNzKSB7IGlmICh0eXBlb2Ygc3VwZXJDbGFzcyAhPT0gJ2Z1bmN0aW9uJyAmJiBzdXBlckNsYXNzICE9PSBudWxsKSB7IHRocm93IG5ldyBUeXBlRXJyb3IoJ1N1cGVyIGV4cHJlc3Npb24gbXVzdCBlaXRoZXIgYmUgbnVsbCBvciBhIGZ1bmN0aW9uLCBub3QgJyArIHR5cGVvZiBzdXBlckNsYXNzKTsgfSBzdWJDbGFzcy5wcm90b3R5cGUgPSBPYmplY3QuY3JlYXRlKHN1cGVyQ2xhc3MgJiYgc3VwZXJDbGFzcy5wcm90b3R5cGUsIHsgY29uc3RydWN0b3I6IHsgdmFsdWU6IHN1YkNsYXNzLCBlbnVtZXJhYmxlOiBmYWxzZSwgd3JpdGFibGU6IHRydWUsIGNvbmZpZ3VyYWJsZTogdHJ1ZSB9IH0pOyBpZiAoc3VwZXJDbGFzcykgT2JqZWN0LnNldFByb3RvdHlwZU9mID8gT2JqZWN0LnNldFByb3RvdHlwZU9mKHN1YkNsYXNzLCBzdXBlckNsYXNzKSA6IHN1YkNsYXNzLl9fcHJvdG9fXyA9IHN1cGVyQ2xhc3M7IH1cblxudmFyIF9yZWFjdCA9IHJlcXVpcmUoJ3JlYWN0Jyk7XG5cbnZhciBfcmVhY3QyID0gX2ludGVyb3BSZXF1aXJlRGVmYXVsdChfcmVhY3QpO1xuXG52YXIgX3B5ZGlvSHR0cEFwaSA9IHJlcXVpcmUoJ3B5ZGlvL2h0dHAvYXBpJyk7XG5cbnZhciBfcHlkaW9IdHRwQXBpMiA9IF9pbnRlcm9wUmVxdWlyZURlZmF1bHQoX3B5ZGlvSHR0cEFwaSk7XG5cbnZhciBfbWF0ZXJpYWxVaSA9IHJlcXVpcmUoJ21hdGVyaWFsLXVpJyk7XG5cbnZhciBfcHlkaW9IdHRwUmVzdEFwaSA9IHJlcXVpcmUoJ3B5ZGlvL2h0dHAvcmVzdC1hcGknKTtcblxudmFyIHN0eWxlcyA9IHtcbiAgICBjaGlwOiB7XG4gICAgICAgIG1hcmdpblJpZ2h0OiA0LFxuICAgICAgICBtYXJnaW5Cb3R0b206IDRcbiAgICB9LFxuICAgIHdyYXBwZXI6IHtcbiAgICAgICAgZGlzcGxheTogJ2ZsZXgnLFxuICAgICAgICBmbGV4V3JhcDogJ3dyYXAnXG4gICAgfSxcbiAgICBvdmVybGF5OiB7XG4gICAgICAgIHBvc2l0aW9uOiAnYWJzb2x1dGUnLFxuICAgICAgICB0b3A6IDAsXG4gICAgICAgIHJpZ2h0OiAwLFxuICAgICAgICBsZWZ0OiAwLFxuICAgICAgICBib3R0b206IDAsXG4gICAgICAgIGJhY2tncm91bmRDb2xvcjogJ3JnYmEoMCwgMCwgMCwgMC4zMyknLFxuICAgICAgICBwYWRkaW5nVG9wOiA3NyxcbiAgICAgICAgekluZGV4OiAxMDBcbiAgICB9XG59O1xuXG52YXIgRGVzdEJhZGdlID0gKGZ1bmN0aW9uIChfUmVhY3QkQ29tcG9uZW50KSB7XG4gICAgX2luaGVyaXRzKERlc3RCYWRnZSwgX1JlYWN0JENvbXBvbmVudCk7XG5cbiAgICBmdW5jdGlvbiBEZXN0QmFkZ2UoKSB7XG4gICAgICAgIF9jbGFzc0NhbGxDaGVjayh0aGlzLCBEZXN0QmFkZ2UpO1xuXG4gICAgICAgIF9nZXQoT2JqZWN0LmdldFByb3RvdHlwZU9mKERlc3RCYWRnZS5wcm90b3R5cGUpLCAnY29uc3RydWN0b3InLCB0aGlzKS5hcHBseSh0aGlzLCBhcmd1bWVudHMpO1xuICAgIH1cblxuICAgIF9jcmVhdGVDbGFzcyhEZXN0QmFkZ2UsIFt7XG4gICAgICAgIGtleTogJ3JlbmRlcicsXG4gICAgICAgIHZhbHVlOiBmdW5jdGlvbiByZW5kZXIoKSB7XG4gICAgICAgICAgICB2YXIgdXNlck9iamVjdCA9IHRoaXMucHJvcHMudXNlcjtcbiAgICAgICAgICAgIHJldHVybiBfcmVhY3QyWydkZWZhdWx0J10uY3JlYXRlRWxlbWVudChcbiAgICAgICAgICAgICAgICAnZGl2JyxcbiAgICAgICAgICAgICAgICB7IGNsYXNzTmFtZTogXCJzaGFyZS1kaWFsb2cgdXNlci1iYWRnZSB1c2VyLXR5cGUtXCIgKyAodXNlck9iamVjdC5nZXRUZW1wb3JhcnkoKSA/IFwidG1wX3VzZXJcIiA6IFwidXNlclwiKSB9LFxuICAgICAgICAgICAgICAgIF9yZWFjdDJbJ2RlZmF1bHQnXS5jcmVhdGVFbGVtZW50KCdzcGFuJywgeyBjbGFzc05hbWU6IFwiYXZhdGFyIGljb24tXCIgKyAodXNlck9iamVjdC5nZXRUZW1wb3JhcnkoKSA/IFwiZW52ZWxvcGVcIiA6IFwidXNlclwiKSB9KSxcbiAgICAgICAgICAgICAgICBfcmVhY3QyWydkZWZhdWx0J10uY3JlYXRlRWxlbWVudChcbiAgICAgICAgICAgICAgICAgICAgJ3NwYW4nLFxuICAgICAgICAgICAgICAgICAgICB7IGNsYXNzTmFtZTogJ3VzZXItYmFkZ2UtbGFiZWwnIH0sXG4gICAgICAgICAgICAgICAgICAgIHVzZXJPYmplY3QuZ2V0RXh0ZW5kZWRMYWJlbCgpIHx8IHVzZXJPYmplY3QuZ2V0TGFiZWwoKVxuICAgICAgICAgICAgICAgIClcbiAgICAgICAgICAgICk7XG4gICAgICAgIH1cbiAgICB9XSk7XG5cbiAgICByZXR1cm4gRGVzdEJhZGdlO1xufSkoX3JlYWN0MlsnZGVmYXVsdCddLkNvbXBvbmVudCk7XG5cbnZhciBVc2VyQ2hpcCA9IChmdW5jdGlvbiAoX1JlYWN0JENvbXBvbmVudDIpIHtcbiAgICBfaW5oZXJpdHMoVXNlckNoaXAsIF9SZWFjdCRDb21wb25lbnQyKTtcblxuICAgIGZ1bmN0aW9uIFVzZXJDaGlwKCkge1xuICAgICAgICBfY2xhc3NDYWxsQ2hlY2sodGhpcywgVXNlckNoaXApO1xuXG4gICAgICAgIF9nZXQoT2JqZWN0LmdldFByb3RvdHlwZU9mKFVzZXJDaGlwLnByb3RvdHlwZSksICdjb25zdHJ1Y3RvcicsIHRoaXMpLmFwcGx5KHRoaXMsIGFyZ3VtZW50cyk7XG4gICAgfVxuXG4gICAgX2NyZWF0ZUNsYXNzKFVzZXJDaGlwLCBbe1xuICAgICAgICBrZXk6ICdyZW5kZXInLFxuICAgICAgICB2YWx1ZTogZnVuY3Rpb24gcmVuZGVyKCkge1xuICAgICAgICAgICAgdmFyIF9wcm9wcyA9IHRoaXMucHJvcHM7XG4gICAgICAgICAgICB2YXIgdXNlciA9IF9wcm9wcy51c2VyO1xuICAgICAgICAgICAgdmFyIG9uUmVtb3ZlID0gX3Byb3BzLm9uUmVtb3ZlO1xuXG4gICAgICAgICAgICB2YXIgdG1wID0gdXNlci5GcmVlVmFsdWU7XG4gICAgICAgICAgICB2YXIgbGFiZWwgPSB1bmRlZmluZWQ7XG4gICAgICAgICAgICBpZiAodG1wKSB7XG4gICAgICAgICAgICAgICAgbGFiZWwgPSB1c2VyLkZyZWVWYWx1ZTtcbiAgICAgICAgICAgIH0gZWxzZSB7XG4gICAgICAgICAgICAgICAgaWYgKHVzZXIuQXR0cmlidXRlcyAmJiB1c2VyLkF0dHJpYnV0ZXNbJ2Rpc3BsYXlOYW1lJ10pIHtcbiAgICAgICAgICAgICAgICAgICAgbGFiZWwgPSB1c2VyLkF0dHJpYnV0ZXNbJ2Rpc3BsYXlOYW1lJ107XG4gICAgICAgICAgICAgICAgfSBlbHNlIHtcbiAgICAgICAgICAgICAgICAgICAgbGFiZWwgPSB1c2VyLkxvZ2luO1xuICAgICAgICAgICAgICAgIH1cbiAgICAgICAgICAgIH1cblxuICAgICAgICAgICAgdmFyIGljb24gPSBfcmVhY3QyWydkZWZhdWx0J10uY3JlYXRlRWxlbWVudChfbWF0ZXJpYWxVaS5Gb250SWNvbiwgeyBjbGFzc05hbWU6IFwiaWNvbi1cIiArICh0bXAgPyBcImVudmVsb3BlXCIgOiBcInVzZXJcIikgfSk7XG4gICAgICAgICAgICB2YXIgY29sb3JzID0gX21hdGVyaWFsVWkuU3R5bGUuY29sb3JzO1xuXG4gICAgICAgICAgICByZXR1cm4gX3JlYWN0MlsnZGVmYXVsdCddLmNyZWF0ZUVsZW1lbnQoXG4gICAgICAgICAgICAgICAgX21hdGVyaWFsVWkuQ2hpcCxcbiAgICAgICAgICAgICAgICB7XG4gICAgICAgICAgICAgICAgICAgIGJhY2tncm91bmRDb2xvcjogdG1wID8gY29sb3JzLmxpZ2h0Qmx1ZTEwMCA6IGNvbG9ycy5ibHVlR3JleTEwMCxcbiAgICAgICAgICAgICAgICAgICAgb25SZXF1ZXN0RGVsZXRlOiBmdW5jdGlvbiAoKSB7XG4gICAgICAgICAgICAgICAgICAgICAgICBvblJlbW92ZSgpO1xuICAgICAgICAgICAgICAgICAgICB9LFxuICAgICAgICAgICAgICAgICAgICBzdHlsZTogc3R5bGVzLmNoaXBcbiAgICAgICAgICAgICAgICB9LFxuICAgICAgICAgICAgICAgIF9yZWFjdDJbJ2RlZmF1bHQnXS5jcmVhdGVFbGVtZW50KF9tYXRlcmlhbFVpLkF2YXRhciwgeyBpY29uOiBpY29uLCBjb2xvcjogdG1wID8gJ3doaXRlJyA6IGNvbG9ycy5ibHVlR3JleTYwMCwgYmFja2dyb3VuZENvbG9yOiB0bXAgPyBjb2xvcnMubGlnaHRCbHVlMzAwIDogY29sb3JzLmJsdWVHcmV5MzAwIH0pLFxuICAgICAgICAgICAgICAgIGxhYmVsXG4gICAgICAgICAgICApO1xuICAgICAgICB9XG4gICAgfV0pO1xuXG4gICAgcmV0dXJuIFVzZXJDaGlwO1xufSkoX3JlYWN0MlsnZGVmYXVsdCddLkNvbXBvbmVudCk7XG5cbnZhciBFbWFpbCA9IChmdW5jdGlvbiAoKSB7XG4gICAgZnVuY3Rpb24gRW1haWwoKSB7XG4gICAgICAgIHZhciBzdWJqZWN0ID0gYXJndW1lbnRzLmxlbmd0aCA8PSAwIHx8IGFyZ3VtZW50c1swXSA9PT0gdW5kZWZpbmVkID8gbnVsbCA6IGFyZ3VtZW50c1swXTtcbiAgICAgICAgdmFyIG1lc3NhZ2UgPSBhcmd1bWVudHMubGVuZ3RoIDw9IDEgfHwgYXJndW1lbnRzWzFdID09PSB1bmRlZmluZWQgPyBudWxsIDogYXJndW1lbnRzWzFdO1xuICAgICAgICB2YXIgbGluayA9IGFyZ3VtZW50cy5sZW5ndGggPD0gMiB8fCBhcmd1bWVudHNbMl0gPT09IHVuZGVmaW5lZCA/IG51bGwgOiBhcmd1bWVudHNbMl07XG5cbiAgICAgICAgX2NsYXNzQ2FsbENoZWNrKHRoaXMsIEVtYWlsKTtcblxuICAgICAgICB0aGlzLl9zdWJqZWN0cyA9IFtdO1xuICAgICAgICB0aGlzLl9tZXNzYWdlcyA9IFtdO1xuICAgICAgICB0aGlzLl90YXJnZXRzID0gW107XG4gICAgICAgIHRoaXMuX2xpbmtzID0gW107XG4gICAgICAgIHRoaXMudGVtcGxhdGVJZCA9IFwiXCI7XG4gICAgICAgIHRoaXMudGVtcGxhdGVEYXRhID0ge307XG4gICAgICAgIGlmIChzdWJqZWN0KSB7XG4gICAgICAgICAgICB0aGlzLl9zdWJqZWN0cy5wdXNoKHN1YmplY3QpO1xuICAgICAgICB9XG4gICAgICAgIGlmIChtZXNzYWdlKSB7XG4gICAgICAgICAgICB0aGlzLl9tZXNzYWdlcy5wdXNoKG1lc3NhZ2UpO1xuICAgICAgICB9XG4gICAgICAgIGlmIChsaW5rKSB7XG4gICAgICAgICAgICB0aGlzLl9saW5rcy5wdXNoKGxpbmspO1xuICAgICAgICB9XG4gICAgfVxuXG4gICAgX2NyZWF0ZUNsYXNzKEVtYWlsLCBbe1xuICAgICAgICBrZXk6ICdhZGRUYXJnZXQnLFxuICAgICAgICB2YWx1ZTogZnVuY3Rpb24gYWRkVGFyZ2V0KHVzZXJPckVtYWlsKSB7XG4gICAgICAgICAgICB2YXIgc3ViamVjdCA9IGFyZ3VtZW50cy5sZW5ndGggPD0gMSB8fCBhcmd1bWVudHNbMV0gPT09IHVuZGVmaW5lZCA/IG51bGwgOiBhcmd1bWVudHNbMV07XG4gICAgICAgICAgICB2YXIgbWVzc2FnZSA9IGFyZ3VtZW50cy5sZW5ndGggPD0gMiB8fCBhcmd1bWVudHNbMl0gPT09IHVuZGVmaW5lZCA/IG51bGwgOiBhcmd1bWVudHNbMl07XG4gICAgICAgICAgICB2YXIgbGluayA9IGFyZ3VtZW50cy5sZW5ndGggPD0gMyB8fCBhcmd1bWVudHNbM10gPT09IHVuZGVmaW5lZCA/IG51bGwgOiBhcmd1bWVudHNbM107XG5cbiAgICAgICAgICAgIHRoaXMuX3RhcmdldHMucHVzaCh1c2VyT3JFbWFpbCk7XG4gICAgICAgICAgICBpZiAoc3ViamVjdCkge1xuICAgICAgICAgICAgICAgIHRoaXMuX3N1YmplY3RzLnB1c2goc3ViamVjdCk7XG4gICAgICAgICAgICB9XG4gICAgICAgICAgICBpZiAobWVzc2FnZSkge1xuICAgICAgICAgICAgICAgIHRoaXMuX21lc3NhZ2VzLnB1c2gobWVzc2FnZSk7XG4gICAgICAgICAgICB9XG4gICAgICAgICAgICBpZiAobGluaykge1xuICAgICAgICAgICAgICAgIHRoaXMuX2xpbmtzLnB1c2gobGluayk7XG4gICAgICAgICAgICB9XG4gICAgICAgIH1cbiAgICB9LCB7XG4gICAgICAgIGtleTogJ3NldFRlbXBsYXRlRGF0YScsXG4gICAgICAgIHZhbHVlOiBmdW5jdGlvbiBzZXRUZW1wbGF0ZURhdGEodGVtcGxhdGVJZCwgdGVtcGxhdGVEYXRhKSB7XG4gICAgICAgICAgICB0aGlzLnRlbXBsYXRlSWQgPSB0ZW1wbGF0ZUlkO1xuICAgICAgICAgICAgdGhpcy50ZW1wbGF0ZURhdGEgPSB0ZW1wbGF0ZURhdGE7XG4gICAgICAgIH1cblxuICAgICAgICAvKipcbiAgICAgICAgICpcbiAgICAgICAgICogQHBhcmFtIHRhcmdldHMgQXJyYXlcbiAgICAgICAgICogQHBhcmFtIHN1YmplY3Qgc3RyaW5nXG4gICAgICAgICAqIEBwYXJhbSB0ZW1wbGF0ZUlkIHN0cmluZ1xuICAgICAgICAgKiBAcGFyYW0gdGVtcGxhdGVEYXRhIE9iamVjdFxuICAgICAgICAgKiBAcmV0dXJuIHtQcm9taXNlfVxuICAgICAgICAgKi9cbiAgICB9LCB7XG4gICAgICAgIGtleTogJ3Bvc3RPbmUnLFxuICAgICAgICB2YWx1ZTogZnVuY3Rpb24gcG9zdE9uZSh0YXJnZXRzLCBzdWJqZWN0LCB0ZW1wbGF0ZUlkLCB0ZW1wbGF0ZURhdGEpIHtcbiAgICAgICAgICAgIHZhciBhcGkgPSBuZXcgX3B5ZGlvSHR0cFJlc3RBcGkuTWFpbGVyU2VydmljZUFwaShfcHlkaW9IdHRwQXBpMlsnZGVmYXVsdCddLmdldFJlc3RDbGllbnQoKSk7XG4gICAgICAgICAgICB2YXIgZW1haWwgPSBuZXcgX3B5ZGlvSHR0cFJlc3RBcGkuTWFpbGVyTWFpbCgpO1xuICAgICAgICAgICAgZW1haWwuU3ViamVjdCA9IHN1YmplY3Q7XG4gICAgICAgICAgICBlbWFpbC5UbyA9IHRhcmdldHMubWFwKGZ1bmN0aW9uICh0KSB7XG4gICAgICAgICAgICAgICAgcmV0dXJuIF9weWRpb0h0dHBSZXN0QXBpLk1haWxlclVzZXIuY29uc3RydWN0RnJvbU9iamVjdCh7IFV1aWQ6IHQgfSk7XG4gICAgICAgICAgICB9KTtcbiAgICAgICAgICAgIGVtYWlsLlRlbXBsYXRlSWQgPSB0ZW1wbGF0ZUlkO1xuICAgICAgICAgICAgZW1haWwuVGVtcGxhdGVEYXRhID0gdGVtcGxhdGVEYXRhO1xuICAgICAgICAgICAgcmV0dXJuIGFwaS5zZW5kKGVtYWlsKTtcbiAgICAgICAgfVxuICAgIH0sIHtcbiAgICAgICAga2V5OiAncG9zdCcsXG4gICAgICAgIHZhbHVlOiBmdW5jdGlvbiBwb3N0KCkge1xuICAgICAgICAgICAgdmFyIF90aGlzID0gdGhpcztcblxuICAgICAgICAgICAgdmFyIGNhbGxiYWNrID0gYXJndW1lbnRzLmxlbmd0aCA8PSAwIHx8IGFyZ3VtZW50c1swXSA9PT0gdW5kZWZpbmVkID8gbnVsbCA6IGFyZ3VtZW50c1swXTtcblxuICAgICAgICAgICAgaWYgKCF0aGlzLl9zdWJqZWN0cy5sZW5ndGggJiYgIXRoaXMudGVtcGxhdGVJZCB8fCAhdGhpcy5fdGFyZ2V0cy5sZW5ndGggfHwgIXRoaXMuX21lc3NhZ2VzLmxlbmd0aCkge1xuICAgICAgICAgICAgICAgIHRocm93IG5ldyBFcnJvcignSW52YWxpZCBkYXRhJyk7XG4gICAgICAgICAgICB9XG4gICAgICAgICAgICB2YXIgdGVtcGxhdGVEYXRhID0gX2V4dGVuZHMoe30sIHRoaXMudGVtcGxhdGVEYXRhKTtcbiAgICAgICAgICAgIHZhciBwcm9tcyA9IFtdO1xuXG4gICAgICAgICAgICBpZiAodGhpcy5fbWVzc2FnZXMubGVuZ3RoID4gMSAmJiB0aGlzLl9tZXNzYWdlcy5sZW5ndGggPT09IHRoaXMuX3RhcmdldHMubGVuZ3RoICYmIHRoaXMuX3N1YmplY3RzLmxlbmd0aCA9PT0gdGhpcy5fdGFyZ2V0cy5sZW5ndGgpIHtcbiAgICAgICAgICAgICAgICAvLyBTZW5kIGFzIG1hbnkgZW1haWxzIGFzIHRhcmdldHMgd2l0aCB0aGVpciBvd24gbWVzc2FnZXNcbiAgICAgICAgICAgICAgICB0aGlzLl90YXJnZXRzLm1hcChmdW5jdGlvbiAodCwgaSkge1xuICAgICAgICAgICAgICAgICAgICB2YXIgc3ViamVjdCA9IF90aGlzLl9zdWJqZWN0c1tpXTtcbiAgICAgICAgICAgICAgICAgICAgdGVtcGxhdGVEYXRhID0gX2V4dGVuZHMoe30sIHRlbXBsYXRlRGF0YSwgeyBNZXNzYWdlOiBfdGhpcy5fbWVzc2FnZXNbaV0gfSk7XG4gICAgICAgICAgICAgICAgICAgIHByb21zLnB1c2goX3RoaXMucG9zdE9uZShbdF0sIHN1YmplY3QsIF90aGlzLnRlbXBsYXRlSWQsIHRlbXBsYXRlRGF0YSkpO1xuICAgICAgICAgICAgICAgIH0pO1xuICAgICAgICAgICAgfSBlbHNlIHtcbiAgICAgICAgICAgICAgICB2YXIgc3ViamVjdCA9IHRoaXMuX3N1YmplY3RzWzBdO1xuICAgICAgICAgICAgICAgIHRlbXBsYXRlRGF0YVsnTWVzc2FnZSddID0gdGhpcy5fbWVzc2FnZXNbMF07XG4gICAgICAgICAgICAgICAgcHJvbXMucHVzaCh0aGlzLnBvc3RPbmUodGhpcy5fdGFyZ2V0cywgc3ViamVjdCwgdGhpcy50ZW1wbGF0ZUlkLCB0ZW1wbGF0ZURhdGEpKTtcbiAgICAgICAgICAgIH1cblxuICAgICAgICAgICAgUHJvbWlzZS5hbGwocHJvbXMpLnRoZW4oZnVuY3Rpb24gKCkge1xuICAgICAgICAgICAgICAgIGNhbGxiYWNrKHRydWUpO1xuICAgICAgICAgICAgfSlbJ2NhdGNoJ10oZnVuY3Rpb24gKGUpIHtcbiAgICAgICAgICAgICAgICBpZiAoZS5yZXNwb25zZSAmJiBlLnJlc3BvbnNlLmJvZHkgJiYgZS5yZXNwb25zZS5ib2R5LlRpdGxlKSB7XG4gICAgICAgICAgICAgICAgICAgIGUgPSBuZXcgRXJyb3IoZS5yZXNwb25zZS5ib2R5LlRpdGxlKTtcbiAgICAgICAgICAgICAgICB9XG4gICAgICAgICAgICAgICAgY2FsbGJhY2soZmFsc2UsIGUpO1xuICAgICAgICAgICAgfSk7XG4gICAgICAgIH1cbiAgICB9XSk7XG5cbiAgICByZXR1cm4gRW1haWw7XG59KSgpO1xuXG52YXIgUGFuZSA9IChmdW5jdGlvbiAoX1JlYWN0JENvbXBvbmVudDMpIHtcbiAgICBfaW5oZXJpdHMoUGFuZSwgX1JlYWN0JENvbXBvbmVudDMpO1xuXG4gICAgZnVuY3Rpb24gUGFuZShwcm9wcykge1xuICAgICAgICBfY2xhc3NDYWxsQ2hlY2sodGhpcywgUGFuZSk7XG5cbiAgICAgICAgaWYgKHByb3BzLnNob3dBZGRyZXNzQm9vayA9PT0gdW5kZWZpbmVkKSB7XG4gICAgICAgICAgICBwcm9wcy5zaG93QWRkcmVzc0Jvb2sgPSB0cnVlO1xuICAgICAgICB9XG4gICAgICAgIF9nZXQoT2JqZWN0LmdldFByb3RvdHlwZU9mKFBhbmUucHJvdG90eXBlKSwgJ2NvbnN0cnVjdG9yJywgdGhpcykuY2FsbCh0aGlzLCBwcm9wcyk7XG4gICAgICAgIHRoaXMuc3RhdGUgPSB7XG4gICAgICAgICAgICB1c2VyczogdGhpcy5wcm9wcy51c2VycyB8fCB7fSxcbiAgICAgICAgICAgIHN1YmplY3Q6IHRoaXMucHJvcHMuc3ViamVjdCxcbiAgICAgICAgICAgIG1lc3NhZ2U6IHRoaXMucHJvcHMubWVzc2FnZSxcbiAgICAgICAgICAgIGVycm9yTWVzc2FnZTogbnVsbCxcbiAgICAgICAgICAgIHBvc3Rpbmc6IGZhbHNlXG4gICAgICAgIH07XG4gICAgfVxuXG4gICAgX2NyZWF0ZUNsYXNzKFBhbmUsIFt7XG4gICAgICAgIGtleTogJ3VwZGF0ZVN1YmplY3QnLFxuICAgICAgICB2YWx1ZTogZnVuY3Rpb24gdXBkYXRlU3ViamVjdChldmVudCkge1xuICAgICAgICAgICAgdGhpcy5zZXRTdGF0ZSh7IHN1YmplY3Q6IGV2ZW50LmN1cnJlbnRUYXJnZXQudmFsdWUgfSk7XG4gICAgICAgIH1cbiAgICB9LCB7XG4gICAgICAgIGtleTogJ3VwZGF0ZU1lc3NhZ2UnLFxuICAgICAgICB2YWx1ZTogZnVuY3Rpb24gdXBkYXRlTWVzc2FnZShldmVudCkge1xuICAgICAgICAgICAgdGhpcy5zZXRTdGF0ZSh7IG1lc3NhZ2U6IGV2ZW50LmN1cnJlbnRUYXJnZXQudmFsdWUgfSk7XG4gICAgICAgIH1cbiAgICB9LCB7XG4gICAgICAgIGtleTogJ2FkZFVzZXInLFxuICAgICAgICB2YWx1ZTogZnVuY3Rpb24gYWRkVXNlcih1c2VyT2JqZWN0KSB7XG4gICAgICAgICAgICB2YXIgdXNlcnMgPSB0aGlzLnN0YXRlLnVzZXJzO1xuXG4gICAgICAgICAgICBpZiAodXNlck9iamVjdC5GcmVlVmFsdWUpIHtcbiAgICAgICAgICAgICAgICB1c2Vyc1t1c2VyT2JqZWN0LkZyZWVWYWx1ZV0gPSB1c2VyT2JqZWN0O1xuICAgICAgICAgICAgfSBlbHNlIGlmICh1c2VyT2JqZWN0LklkbVVzZXIpIHtcbiAgICAgICAgICAgICAgICB1c2Vyc1t1c2VyT2JqZWN0LklkbVVzZXIuTG9naW5dID0gdXNlck9iamVjdC5JZG1Vc2VyO1xuICAgICAgICAgICAgfVxuICAgICAgICAgICAgdGhpcy5zZXRTdGF0ZSh7IHVzZXJzOiB1c2VycywgZXJyb3JNZXNzYWdlOiBudWxsIH0pO1xuICAgICAgICB9XG4gICAgfSwge1xuICAgICAgICBrZXk6ICdyZW1vdmVVc2VyJyxcbiAgICAgICAgdmFsdWU6IGZ1bmN0aW9uIHJlbW92ZVVzZXIodXNlcklkKSB7XG4gICAgICAgICAgICB2YXIgdXNlcnMgPSB0aGlzLnN0YXRlLnVzZXJzO1xuXG4gICAgICAgICAgICBkZWxldGUgdXNlcnNbdXNlcklkXTtcbiAgICAgICAgICAgIHRoaXMuc2V0U3RhdGUoeyB1c2VyczogdXNlcnMgfSk7XG4gICAgICAgIH1cbiAgICB9LCB7XG4gICAgICAgIGtleTogJ2dldE1lc3NhZ2UnLFxuICAgICAgICB2YWx1ZTogZnVuY3Rpb24gZ2V0TWVzc2FnZShtZXNzYWdlSWQpIHtcbiAgICAgICAgICAgIHZhciBuYW1lU3BhY2UgPSBhcmd1bWVudHMubGVuZ3RoIDw9IDEgfHwgYXJndW1lbnRzWzFdID09PSB1bmRlZmluZWQgPyB1bmRlZmluZWQgOiBhcmd1bWVudHNbMV07XG5cbiAgICAgICAgICAgIHRyeSB7XG4gICAgICAgICAgICAgICAgaWYgKG5hbWVTcGFjZSA9PT0gdW5kZWZpbmVkKSB7XG4gICAgICAgICAgICAgICAgICAgIG5hbWVTcGFjZSA9ICdjb3JlLm1haWxlcic7XG4gICAgICAgICAgICAgICAgfVxuICAgICAgICAgICAgICAgIGlmIChuYW1lU3BhY2UpIHtcbiAgICAgICAgICAgICAgICAgICAgbmFtZVNwYWNlICs9IFwiLlwiO1xuICAgICAgICAgICAgICAgIH1cbiAgICAgICAgICAgICAgICByZXR1cm4gcHlkaW8uTWVzc2FnZUhhc2hbbmFtZVNwYWNlICsgbWVzc2FnZUlkXTtcbiAgICAgICAgICAgIH0gY2F0Y2ggKGUpIHtcbiAgICAgICAgICAgICAgICByZXR1cm4gbWVzc2FnZUlkO1xuICAgICAgICAgICAgfVxuICAgICAgICB9XG4gICAgfSwge1xuICAgICAgICBrZXk6ICdwb3N0RW1haWwnLFxuICAgICAgICB2YWx1ZTogZnVuY3Rpb24gcG9zdEVtYWlsKCkge1xuICAgICAgICAgICAgdmFyIF90aGlzMiA9IHRoaXM7XG5cbiAgICAgICAgICAgIHZhciByZXBvc3QgPSBhcmd1bWVudHMubGVuZ3RoIDw9IDAgfHwgYXJndW1lbnRzWzBdID09PSB1bmRlZmluZWQgPyBmYWxzZSA6IGFyZ3VtZW50c1swXTtcbiAgICAgICAgICAgIHZhciBfc3RhdGUgPSB0aGlzLnN0YXRlO1xuICAgICAgICAgICAgdmFyIHVzZXJzID0gX3N0YXRlLnVzZXJzO1xuICAgICAgICAgICAgdmFyIHN1YmplY3QgPSBfc3RhdGUuc3ViamVjdDtcbiAgICAgICAgICAgIHZhciBtZXNzYWdlID0gX3N0YXRlLm1lc3NhZ2U7XG5cbiAgICAgICAgICAgIGlmICghcmVwb3N0ICYmIHRoaXMucmVmcy5jb21wbGV0ZXIgJiYgdGhpcy5yZWZzLmNvbXBsZXRlci5nZXRQZW5kaW5nU2VhcmNoVGV4dCAmJiB0aGlzLnJlZnMuY29tcGxldGVyLmdldFBlbmRpbmdTZWFyY2hUZXh0KCkpIHtcbiAgICAgICAgICAgICAgICB0aGlzLnJlZnMuY29tcGxldGVyLm9uQ29tcGxldGVyUmVxdWVzdCh0aGlzLnJlZnMuY29tcGxldGVyLmdldFBlbmRpbmdTZWFyY2hUZXh0KCksIC0xKTtcbiAgICAgICAgICAgICAgICBzZXRUaW1lb3V0KGZ1bmN0aW9uICgpIHtcbiAgICAgICAgICAgICAgICAgICAgcmV0dXJuIF90aGlzMi5wb3N0RW1haWwodHJ1ZSk7XG4gICAgICAgICAgICAgICAgfSwgNTAwKTtcbiAgICAgICAgICAgICAgICByZXR1cm47XG4gICAgICAgICAgICB9XG4gICAgICAgICAgICBpZiAoIU9iamVjdC5rZXlzKHVzZXJzKS5sZW5ndGgpIHtcbiAgICAgICAgICAgICAgICB0aGlzLnNldFN0YXRlKHsgZXJyb3JNZXNzYWdlOiB0aGlzLmdldE1lc3NhZ2UoMikgfSk7XG4gICAgICAgICAgICAgICAgcmV0dXJuO1xuICAgICAgICAgICAgfVxuICAgICAgICAgICAgdmFyIF9wcm9wczIgPSB0aGlzLnByb3BzO1xuICAgICAgICAgICAgdmFyIGxpbmsgPSBfcHJvcHMyLmxpbms7XG4gICAgICAgICAgICB2YXIgdGVtcGxhdGVJZCA9IF9wcm9wczIudGVtcGxhdGVJZDtcbiAgICAgICAgICAgIHZhciB0ZW1wbGF0ZURhdGEgPSBfcHJvcHMyLnRlbXBsYXRlRGF0YTtcblxuICAgICAgICAgICAgdmFyIGNhbGxiYWNrID0gZnVuY3Rpb24gY2FsbGJhY2socmVzLCBlcnIpIHtcbiAgICAgICAgICAgICAgICBfdGhpczIuc2V0U3RhdGUoeyBwb3N0aW5nOiBmYWxzZSB9KTtcbiAgICAgICAgICAgICAgICBpZiAocmVzKSB7XG4gICAgICAgICAgICAgICAgICAgIF90aGlzMi5wcm9wcy5weWRpby5VSS5kaXNwbGF5TWVzc2FnZSgnU1VDQ0VTUycsIF90aGlzMi5wcm9wcy5weWRpby5NZXNzYWdlSGFzaFtcImNvcmUubWFpbGVyLjFcIl0ucmVwbGFjZSgnJXMnLCBPYmplY3Qua2V5cyh1c2VycykubGVuZ3RoKSk7XG4gICAgICAgICAgICAgICAgICAgIGlmIChfdGhpczIucHJvcHMub25EaXNtaXNzKSB7XG4gICAgICAgICAgICAgICAgICAgICAgICBfdGhpczIucHJvcHMub25EaXNtaXNzKCk7XG4gICAgICAgICAgICAgICAgICAgIH0gZWxzZSB7XG4gICAgICAgICAgICAgICAgICAgICAgICBfdGhpczIuc2V0U3RhdGUoeyB1c2Vyczoge30sIHN1YmplY3Q6ICcnLCBtZXNzYWdlOiAnJyB9KTtcbiAgICAgICAgICAgICAgICAgICAgfVxuICAgICAgICAgICAgICAgIH0gZWxzZSB7XG4gICAgICAgICAgICAgICAgICAgIF90aGlzMi5wcm9wcy5weWRpby5VSS5kaXNwbGF5TWVzc2FnZSgnRVJST1InLCBfdGhpczIucHJvcHMucHlkaW8uTWVzc2FnZUhhc2hbXCJjb3JlLm1haWxlci5zZW5kZXIuZXJyb3JcIl0gKyAoZXJyID8gJyA6ICcgKyBlcnIubWVzc2FnZSA6ICcnKSk7XG4gICAgICAgICAgICAgICAgfVxuICAgICAgICAgICAgfTtcbiAgICAgICAgICAgIHRoaXMuc2V0U3RhdGUoeyBwb3N0aW5nOiB0cnVlIH0pO1xuXG4gICAgICAgICAgICBpZiAodGhpcy5wcm9wcy5wcm9jZXNzUG9zdCkge1xuICAgICAgICAgICAgICAgIHRoaXMucHJvcHMucHJvY2Vzc1Bvc3QoRW1haWwsIHVzZXJzLCBzdWJqZWN0LCBtZXNzYWdlLCBsaW5rLCBjYWxsYmFjayk7XG4gICAgICAgICAgICAgICAgcmV0dXJuO1xuICAgICAgICAgICAgfVxuXG4gICAgICAgICAgICB2YXIgZW1haWwgPSBuZXcgRW1haWwoc3ViamVjdCwgbWVzc2FnZSwgbGluayB8fCBudWxsKTtcbiAgICAgICAgICAgIE9iamVjdC5rZXlzKHVzZXJzKS5mb3JFYWNoKGZ1bmN0aW9uIChrKSB7XG4gICAgICAgICAgICAgICAgZW1haWwuYWRkVGFyZ2V0KGspO1xuICAgICAgICAgICAgfSk7XG4gICAgICAgICAgICBpZiAodGVtcGxhdGVJZCkge1xuICAgICAgICAgICAgICAgIGVtYWlsLnNldFRlbXBsYXRlRGF0YSh0ZW1wbGF0ZUlkLCB0ZW1wbGF0ZURhdGEpO1xuICAgICAgICAgICAgfVxuICAgICAgICAgICAgZW1haWwucG9zdChjYWxsYmFjayk7XG4gICAgICAgIH1cbiAgICB9LCB7XG4gICAgICAgIGtleTogJ3JlbmRlcicsXG4gICAgICAgIHZhbHVlOiBmdW5jdGlvbiByZW5kZXIoKSB7XG4gICAgICAgICAgICB2YXIgX3RoaXM0ID0gdGhpcztcblxuICAgICAgICAgICAgdmFyIF9zdGF0ZTIgPSB0aGlzLnN0YXRlO1xuICAgICAgICAgICAgdmFyIHVzZXJzID0gX3N0YXRlMi51c2VycztcbiAgICAgICAgICAgIHZhciBwb3N0aW5nID0gX3N0YXRlMi5wb3N0aW5nO1xuICAgICAgICAgICAgdmFyIGVycm9yTWVzc2FnZSA9IF9zdGF0ZTIuZXJyb3JNZXNzYWdlO1xuICAgICAgICAgICAgdmFyIHN1YmplY3QgPSBfc3RhdGUyLnN1YmplY3Q7XG4gICAgICAgICAgICB2YXIgbWVzc2FnZSA9IF9zdGF0ZTIubWVzc2FnZTtcblxuICAgICAgICAgICAgdmFyIGNsYXNzTmFtZSA9IFt0aGlzLnByb3BzLmNsYXNzTmFtZSwgXCJyZWFjdC1tYWlsZXJcIiwgXCJyZXNldC1weWRpby1mb3Jtc1wiXS5qb2luKFwiIFwiKTtcbiAgICAgICAgICAgIHZhciB1c2Vyc0NoaXBzID0gT2JqZWN0LmtleXModXNlcnMpLm1hcCgoZnVuY3Rpb24gKHVJZCkge1xuICAgICAgICAgICAgICAgIHZhciBfdGhpczMgPSB0aGlzO1xuXG4gICAgICAgICAgICAgICAgcmV0dXJuIF9yZWFjdDJbJ2RlZmF1bHQnXS5jcmVhdGVFbGVtZW50KFVzZXJDaGlwLCB7IGtleTogdUlkLCB1c2VyOiB1c2Vyc1t1SWRdLCBvblJlbW92ZTogZnVuY3Rpb24gKCkge1xuICAgICAgICAgICAgICAgICAgICAgICAgX3RoaXMzLnJlbW92ZVVzZXIodUlkKTtcbiAgICAgICAgICAgICAgICAgICAgfSB9KTtcbiAgICAgICAgICAgIH0pLmJpbmQodGhpcykpO1xuICAgICAgICAgICAgdmFyIGVycm9yRGl2ID0gdW5kZWZpbmVkO1xuICAgICAgICAgICAgaWYgKGVycm9yTWVzc2FnZSkge1xuICAgICAgICAgICAgICAgIGVycm9yRGl2ID0gX3JlYWN0MlsnZGVmYXVsdCddLmNyZWF0ZUVsZW1lbnQoXG4gICAgICAgICAgICAgICAgICAgICdkaXYnLFxuICAgICAgICAgICAgICAgICAgICB7IHN0eWxlOiB7IHBhZGRpbmc6ICcxMHB4IDIwcHgnLCBjb2xvcjogJ3JlZCcgfSB9LFxuICAgICAgICAgICAgICAgICAgICBlcnJvck1lc3NhZ2VcbiAgICAgICAgICAgICAgICApO1xuICAgICAgICAgICAgfVxuICAgICAgICAgICAgdmFyIHN0eWxlID0gX2V4dGVuZHMoe1xuICAgICAgICAgICAgICAgIG1hcmdpbjogdGhpcy5wcm9wcy51bmlxdWVVc2VyU3R5bGUgPyAwIDogOFxuICAgICAgICAgICAgfSwgdGhpcy5wcm9wcy5zdHlsZSk7XG4gICAgICAgICAgICB2YXIgY29udGVudCA9IF9yZWFjdDJbJ2RlZmF1bHQnXS5jcmVhdGVFbGVtZW50KFxuICAgICAgICAgICAgICAgIF9tYXRlcmlhbFVpLlBhcGVyLFxuICAgICAgICAgICAgICAgIHsgekRlcHRoOiB0aGlzLnByb3BzLnpEZXB0aCAhPT0gdW5kZWZpbmVkID8gdGhpcy5wcm9wcy56RGVwdGggOiAyLCBjbGFzc05hbWU6IGNsYXNzTmFtZSwgc3R5bGU6IHN0eWxlIH0sXG4gICAgICAgICAgICAgICAgX3JlYWN0MlsnZGVmYXVsdCddLmNyZWF0ZUVsZW1lbnQoXG4gICAgICAgICAgICAgICAgICAgICdoMycsXG4gICAgICAgICAgICAgICAgICAgIHsgc3R5bGU6IF9leHRlbmRzKHsgcGFkZGluZzogMjAsIGNvbG9yOiAncmdiYSgwLDAsMCwwLjg3KScsIGZvbnRTaXplOiAyNSwgbWFyZ2luQm90dG9tOiAwLCBwYWRkaW5nQm90dG9tOiAxMCB9LCB0aGlzLnByb3BzLnRpdGxlU3R5bGUpIH0sXG4gICAgICAgICAgICAgICAgICAgIHRoaXMucHJvcHMucGFuZWxUaXRsZVxuICAgICAgICAgICAgICAgICksXG4gICAgICAgICAgICAgICAgZXJyb3JEaXYsXG4gICAgICAgICAgICAgICAgdGhpcy5wcm9wcy5hZGRpdGlvbmFsUGFuZVRvcCxcbiAgICAgICAgICAgICAgICAhdGhpcy5wcm9wcy51bmlxdWVVc2VyU3R5bGUgJiYgX3JlYWN0MlsnZGVmYXVsdCddLmNyZWF0ZUVsZW1lbnQoXG4gICAgICAgICAgICAgICAgICAgICdkaXYnLFxuICAgICAgICAgICAgICAgICAgICB7IGNsYXNzTmFtZTogJ3VzZXJzLWJsb2NrJywgc3R5bGU6IF9leHRlbmRzKHsgcGFkZGluZzogJzAgMjBweCcgfSwgdGhpcy5wcm9wcy51c2Vyc0Jsb2NrU3R5bGUpIH0sXG4gICAgICAgICAgICAgICAgICAgIF9yZWFjdDJbJ2RlZmF1bHQnXS5jcmVhdGVFbGVtZW50KFB5ZGlvQ29tcG9uZW50cy5Vc2Vyc0NvbXBsZXRlciwge1xuICAgICAgICAgICAgICAgICAgICAgICAgcmVmOiAnY29tcGxldGVyJyxcbiAgICAgICAgICAgICAgICAgICAgICAgIGZpZWxkTGFiZWw6IHRoaXMuZ2V0TWVzc2FnZSgnOCcpLFxuICAgICAgICAgICAgICAgICAgICAgICAgdXNlcnNPbmx5OiB0cnVlLFxuICAgICAgICAgICAgICAgICAgICAgICAgZXhpc3RpbmdPbmx5OiB0cnVlLFxuICAgICAgICAgICAgICAgICAgICAgICAgZnJlZVZhbHVlQWxsb3dlZDogdHJ1ZSxcbiAgICAgICAgICAgICAgICAgICAgICAgIG9uVmFsdWVTZWxlY3RlZDogdGhpcy5hZGRVc2VyLmJpbmQodGhpcyksXG4gICAgICAgICAgICAgICAgICAgICAgICBleGNsdWRlczogT2JqZWN0LmtleXModXNlcnMpLFxuICAgICAgICAgICAgICAgICAgICAgICAgcmVuZGVyU3VnZ2VzdGlvbjogZnVuY3Rpb24gKHVzZXJPYmplY3QpIHtcbiAgICAgICAgICAgICAgICAgICAgICAgICAgICByZXR1cm4gX3JlYWN0MlsnZGVmYXVsdCddLmNyZWF0ZUVsZW1lbnQoRGVzdEJhZGdlLCB7IHVzZXI6IHVzZXJPYmplY3QgfSk7XG4gICAgICAgICAgICAgICAgICAgICAgICB9LFxuICAgICAgICAgICAgICAgICAgICAgICAgcHlkaW86IHB5ZGlvLFxuICAgICAgICAgICAgICAgICAgICAgICAgc2hvd0FkZHJlc3NCb29rOiB0aGlzLnByb3BzLnNob3dBZGRyZXNzQm9vayxcbiAgICAgICAgICAgICAgICAgICAgICAgIHVuZGVybGluZUhpZGU6IHRydWVcbiAgICAgICAgICAgICAgICAgICAgfSksXG4gICAgICAgICAgICAgICAgICAgIF9yZWFjdDJbJ2RlZmF1bHQnXS5jcmVhdGVFbGVtZW50KFxuICAgICAgICAgICAgICAgICAgICAgICAgJ2RpdicsXG4gICAgICAgICAgICAgICAgICAgICAgICB7IHN0eWxlOiBzdHlsZXMud3JhcHBlciB9LFxuICAgICAgICAgICAgICAgICAgICAgICAgdXNlcnNDaGlwc1xuICAgICAgICAgICAgICAgICAgICApXG4gICAgICAgICAgICAgICAgKSxcbiAgICAgICAgICAgICAgICAhdGhpcy5wcm9wcy51bmlxdWVVc2VyU3R5bGUgJiYgX3JlYWN0MlsnZGVmYXVsdCddLmNyZWF0ZUVsZW1lbnQoX21hdGVyaWFsVWkuRGl2aWRlciwgbnVsbCksXG4gICAgICAgICAgICAgICAgIXRoaXMucHJvcHMudGVtcGxhdGVJZCAmJiBfcmVhY3QyWydkZWZhdWx0J10uY3JlYXRlRWxlbWVudChcbiAgICAgICAgICAgICAgICAgICAgJ2RpdicsXG4gICAgICAgICAgICAgICAgICAgIHsgc3R5bGU6IHsgcGFkZGluZzogJzAgMjBweCcgfSB9LFxuICAgICAgICAgICAgICAgICAgICBfcmVhY3QyWydkZWZhdWx0J10uY3JlYXRlRWxlbWVudChfbWF0ZXJpYWxVaS5UZXh0RmllbGQsIHsgZnVsbFdpZHRoOiB0cnVlLCB1bmRlcmxpbmVTaG93OiBmYWxzZSwgZmxvYXRpbmdMYWJlbFRleHQ6IHRoaXMuZ2V0TWVzc2FnZSgnNicpLCB2YWx1ZTogc3ViamVjdCwgb25DaGFuZ2U6IHRoaXMudXBkYXRlU3ViamVjdC5iaW5kKHRoaXMpIH0pXG4gICAgICAgICAgICAgICAgKSxcbiAgICAgICAgICAgICAgICAhdGhpcy5wcm9wcy50ZW1wbGF0ZUlkICYmIF9yZWFjdDJbJ2RlZmF1bHQnXS5jcmVhdGVFbGVtZW50KF9tYXRlcmlhbFVpLkRpdmlkZXIsIG51bGwpLFxuICAgICAgICAgICAgICAgIF9yZWFjdDJbJ2RlZmF1bHQnXS5jcmVhdGVFbGVtZW50KFxuICAgICAgICAgICAgICAgICAgICAnZGl2JyxcbiAgICAgICAgICAgICAgICAgICAgeyBzdHlsZTogX2V4dGVuZHMoeyBwYWRkaW5nOiAnMCAyMHB4JyB9LCB0aGlzLnByb3BzLm1lc3NhZ2VCbG9ja1N0eWxlKSB9LFxuICAgICAgICAgICAgICAgICAgICBfcmVhY3QyWydkZWZhdWx0J10uY3JlYXRlRWxlbWVudChfbWF0ZXJpYWxVaS5UZXh0RmllbGQsIHtcbiAgICAgICAgICAgICAgICAgICAgICAgIGZ1bGxXaWR0aDogdHJ1ZSxcbiAgICAgICAgICAgICAgICAgICAgICAgIHVuZGVybGluZVNob3c6IGZhbHNlLFxuICAgICAgICAgICAgICAgICAgICAgICAgZmxvYXRpbmdMYWJlbFRleHQ6IHRoaXMuZ2V0TWVzc2FnZSgnNycpLFxuICAgICAgICAgICAgICAgICAgICAgICAgdmFsdWU6IG1lc3NhZ2UsXG4gICAgICAgICAgICAgICAgICAgICAgICBtdWx0aUxpbmU6IHRydWUsXG4gICAgICAgICAgICAgICAgICAgICAgICBvbkNoYW5nZTogdGhpcy51cGRhdGVNZXNzYWdlLmJpbmQodGhpcyksXG4gICAgICAgICAgICAgICAgICAgICAgICByb3dzTWF4OiA2XG4gICAgICAgICAgICAgICAgICAgIH0pXG4gICAgICAgICAgICAgICAgKSxcbiAgICAgICAgICAgICAgICB0aGlzLnByb3BzLmFkZGl0aW9uYWxQYW5lQm90dG9tLFxuICAgICAgICAgICAgICAgIF9yZWFjdDJbJ2RlZmF1bHQnXS5jcmVhdGVFbGVtZW50KF9tYXRlcmlhbFVpLkRpdmlkZXIsIG51bGwpLFxuICAgICAgICAgICAgICAgIF9yZWFjdDJbJ2RlZmF1bHQnXS5jcmVhdGVFbGVtZW50KFxuICAgICAgICAgICAgICAgICAgICAnZGl2JyxcbiAgICAgICAgICAgICAgICAgICAgeyBzdHlsZTogeyB0ZXh0QWxpZ246ICdyaWdodCcsIHBhZGRpbmc6ICc4cHggMjBweCcgfSB9LFxuICAgICAgICAgICAgICAgICAgICB0aGlzLnByb3BzLm9uRGlzbWlzcyAmJiBfcmVhY3QyWydkZWZhdWx0J10uY3JlYXRlRWxlbWVudChfbWF0ZXJpYWxVaS5GbGF0QnV0dG9uLCB7IGxhYmVsOiB0aGlzLmdldE1lc3NhZ2UoJzU0JywgJycpLCBvblRvdWNoVGFwOiB0aGlzLnByb3BzLm9uRGlzbWlzcyB9KSxcbiAgICAgICAgICAgICAgICAgICAgIXRoaXMucHJvcHMub25EaXNtaXNzICYmIF9yZWFjdDJbJ2RlZmF1bHQnXS5jcmVhdGVFbGVtZW50KF9tYXRlcmlhbFVpLkZsYXRCdXR0b24sIHsgbGFiZWw6IHRoaXMuZ2V0TWVzc2FnZSgnMjE2JywgJycpLCBvblRvdWNoVGFwOiBmdW5jdGlvbiAoKSB7XG4gICAgICAgICAgICAgICAgICAgICAgICAgICAgX3RoaXM0LnNldFN0YXRlKHsgdXNlcnM6IHt9LCBzdWJqZWN0OiAnJywgbWVzc2FnZTogJycgfSk7XG4gICAgICAgICAgICAgICAgICAgICAgICB9IH0pLFxuICAgICAgICAgICAgICAgICAgICBfcmVhY3QyWydkZWZhdWx0J10uY3JlYXRlRWxlbWVudChfbWF0ZXJpYWxVaS5GbGF0QnV0dG9uLCB7IGRpc2FibGVkOiBwb3N0aW5nLCBwcmltYXJ5OiB0cnVlLCBsYWJlbDogdGhpcy5nZXRNZXNzYWdlKCc3NycsICcnKSwgb25Ub3VjaFRhcDogZnVuY3Rpb24gKGUpIHtcbiAgICAgICAgICAgICAgICAgICAgICAgICAgICByZXR1cm4gX3RoaXM0LnBvc3RFbWFpbCgpO1xuICAgICAgICAgICAgICAgICAgICAgICAgfSB9KVxuICAgICAgICAgICAgICAgIClcbiAgICAgICAgICAgICk7XG4gICAgICAgICAgICBpZiAodGhpcy5wcm9wcy5vdmVybGF5KSB7XG4gICAgICAgICAgICAgICAgcmV0dXJuIF9yZWFjdDJbJ2RlZmF1bHQnXS5jcmVhdGVFbGVtZW50KFxuICAgICAgICAgICAgICAgICAgICAnZGl2JyxcbiAgICAgICAgICAgICAgICAgICAgeyBzdHlsZTogc3R5bGVzLm92ZXJsYXkgfSxcbiAgICAgICAgICAgICAgICAgICAgY29udGVudFxuICAgICAgICAgICAgICAgICk7XG4gICAgICAgICAgICB9IGVsc2Uge1xuICAgICAgICAgICAgICAgIHJldHVybiBjb250ZW50O1xuICAgICAgICAgICAgfVxuICAgICAgICB9XG4gICAgfV0pO1xuXG4gICAgcmV0dXJuIFBhbmU7XG59KShfcmVhY3QyWydkZWZhdWx0J10uQ29tcG9uZW50KTtcblxuUGFuZS5Qcm9wVHlwZXMgPSB7XG4gICAgbWVzc2FnZTogX3JlYWN0MlsnZGVmYXVsdCddLlByb3BUeXBlcy5zdHJpbmcsXG4gICAgc3ViamVjdDogX3JlYWN0MlsnZGVmYXVsdCddLlByb3BUeXBlcy5zdHJpbmcsXG4gICAgdGVtcGxhdGVJZDogX3JlYWN0MlsnZGVmYXVsdCddLlByb3BUeXBlcy5zdHJpbmcsXG4gICAgdGVtcGxhdGVEYXRhOiBfcmVhY3QyWydkZWZhdWx0J10uUHJvcFR5cGVzLm9iamVjdCxcbiAgICBsaW5rOiBfcmVhY3QyWydkZWZhdWx0J10uUHJvcFR5cGVzLnN0cmluZyxcbiAgICBvbkRpc21pc3M6IF9yZWFjdDJbJ2RlZmF1bHQnXS5Qcm9wVHlwZXMuZnVuYyxcbiAgICBjbGFzc05hbWU6IF9yZWFjdDJbJ2RlZmF1bHQnXS5Qcm9wVHlwZXMuc3RyaW5nLFxuICAgIG92ZXJsYXk6IF9yZWFjdDJbJ2RlZmF1bHQnXS5Qcm9wVHlwZXMuYm9vbCxcbiAgICB1bmlxdWVVc2VyU3R5bGU6IF9yZWFjdDJbJ2RlZmF1bHQnXS5Qcm9wVHlwZXMuYm9vbCxcbiAgICB1c2VyczogX3JlYWN0MlsnZGVmYXVsdCddLlByb3BUeXBlcy5vYmplY3QsXG4gICAgcGFuZWxUaXRsZTogX3JlYWN0MlsnZGVmYXVsdCddLlByb3BUeXBlcy5zdHJpbmcsXG4gICAgekRlcHRoOiBfcmVhY3QyWydkZWZhdWx0J10uUHJvcFR5cGVzLm51bWJlcixcbiAgICBzaG93QWRkcmVzc0Jvb2s6IF9yZWFjdDJbJ2RlZmF1bHQnXS5Qcm9wVHlwZXMuYm9vbCxcbiAgICBwcm9jZXNzUG9zdDogX3JlYWN0MlsnZGVmYXVsdCddLlByb3BUeXBlcy5mdW5jLFxuICAgIGFkZGl0aW9uYWxQYW5lVG9wOiBfcmVhY3QyWydkZWZhdWx0J10uUHJvcFR5cGVzLmluc3RhbmNlT2YoX3JlYWN0MlsnZGVmYXVsdCddLkNvbXBvbmVudCksXG4gICAgYWRkaXRpb25hbFBhbmVCb3R0b206IF9yZWFjdDJbJ2RlZmF1bHQnXS5Qcm9wVHlwZXMuaW5zdGFuY2VPZihfcmVhY3QyWydkZWZhdWx0J10uQ29tcG9uZW50KVxufTtcblxudmFyIFByZWZlcmVuY2VzUGFuZWwgPSAoZnVuY3Rpb24gKF9SZWFjdCRDb21wb25lbnQ0KSB7XG4gICAgX2luaGVyaXRzKFByZWZlcmVuY2VzUGFuZWwsIF9SZWFjdCRDb21wb25lbnQ0KTtcblxuICAgIGZ1bmN0aW9uIFByZWZlcmVuY2VzUGFuZWwoKSB7XG4gICAgICAgIF9jbGFzc0NhbGxDaGVjayh0aGlzLCBQcmVmZXJlbmNlc1BhbmVsKTtcblxuICAgICAgICBfZ2V0KE9iamVjdC5nZXRQcm90b3R5cGVPZihQcmVmZXJlbmNlc1BhbmVsLnByb3RvdHlwZSksICdjb25zdHJ1Y3RvcicsIHRoaXMpLmFwcGx5KHRoaXMsIGFyZ3VtZW50cyk7XG4gICAgfVxuXG4gICAgX2NyZWF0ZUNsYXNzKFByZWZlcmVuY2VzUGFuZWwsIFt7XG4gICAgICAgIGtleTogJ3JlbmRlcicsXG4gICAgICAgIHZhbHVlOiBmdW5jdGlvbiByZW5kZXIoKSB7XG4gICAgICAgICAgICByZXR1cm4gX3JlYWN0MlsnZGVmYXVsdCddLmNyZWF0ZUVsZW1lbnQoXG4gICAgICAgICAgICAgICAgJ2RpdicsXG4gICAgICAgICAgICAgICAgbnVsbCxcbiAgICAgICAgICAgICAgICAnUHJlZmVyZW5jZXMgUGFuZWwnXG4gICAgICAgICAgICApO1xuICAgICAgICB9XG4gICAgfV0pO1xuXG4gICAgcmV0dXJuIFByZWZlcmVuY2VzUGFuZWw7XG59KShfcmVhY3QyWydkZWZhdWx0J10uQ29tcG9uZW50KTtcblxuZXhwb3J0cy5QYW5lID0gUGFuZTtcbmV4cG9ydHMuUHJlZmVyZW5jZXNQYW5lbCA9IFByZWZlcmVuY2VzUGFuZWw7XG4iLCJcInVzZSBzdHJpY3RcIjtcblxuT2JqZWN0LmRlZmluZVByb3BlcnR5KGV4cG9ydHMsIFwiX19lc01vZHVsZVwiLCB7XG4gIHZhbHVlOiB0cnVlXG59KTtcblxudmFyIF9jb21wb25lbnRzID0gcmVxdWlyZShcIi4vY29tcG9uZW50c1wiKTtcblxuZXhwb3J0cy5QYW5lID0gX2NvbXBvbmVudHMuUGFuZTtcbmV4cG9ydHMuUHJlZmVyZW5jZXNQYW5lbCA9IF9jb21wb25lbnRzLlByZWZlcmVuY2VzUGFuZWw7XG4iXX0=
