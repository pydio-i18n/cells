'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }(); /**
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      * Pydio Cells Rest API
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      *
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      * The version of the OpenAPI document: 4.0
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      * 
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      *
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      * https://openapi-generator.tech
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      * Do not edit the class manually.
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      *
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      */

var _ApiClient = require('../ApiClient');

var _ApiClient2 = _interopRequireDefault(_ApiClient);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

/**
 * The InstallTLSLetsEncrypt model module.
 * @module model/InstallTLSLetsEncrypt
 * @version 4.0
 */
var InstallTLSLetsEncrypt = function () {
  /**
   * Constructs a new <code>InstallTLSLetsEncrypt</code>.
   * @alias module:model/InstallTLSLetsEncrypt
   */
  function InstallTLSLetsEncrypt() {
    _classCallCheck(this, InstallTLSLetsEncrypt);

    InstallTLSLetsEncrypt.initialize(this);
  }

  /**
   * Initializes the fields of this object.
   * This method is used by the constructors of any subclasses, in order to implement multiple inheritance (mix-ins).
   * Only for internal use.
   */


  _createClass(InstallTLSLetsEncrypt, null, [{
    key: 'initialize',
    value: function initialize(obj) {}

    /**
     * Constructs a <code>InstallTLSLetsEncrypt</code> from a plain JavaScript object, optionally creating a new instance.
     * Copies all relevant properties from <code>data</code> to <code>obj</code> if supplied or a new instance if not.
     * @param {Object} data The plain JavaScript object bearing properties of interest.
     * @param {module:model/InstallTLSLetsEncrypt} obj Optional instance to populate.
     * @return {module:model/InstallTLSLetsEncrypt} The populated <code>InstallTLSLetsEncrypt</code> instance.
     */

  }, {
    key: 'constructFromObject',
    value: function constructFromObject(data, obj) {
      if (data) {
        obj = obj || new InstallTLSLetsEncrypt();

        if (data.hasOwnProperty('AcceptEULA')) {
          obj['AcceptEULA'] = _ApiClient2.default.convertToType(data['AcceptEULA'], 'Boolean');
        }
        if (data.hasOwnProperty('Email')) {
          obj['Email'] = _ApiClient2.default.convertToType(data['Email'], 'String');
        }
        if (data.hasOwnProperty('StagingCA')) {
          obj['StagingCA'] = _ApiClient2.default.convertToType(data['StagingCA'], 'Boolean');
        }
      }
      return obj;
    }
  }]);

  return InstallTLSLetsEncrypt;
}();

/**
 * @member {Boolean} AcceptEULA
 */


InstallTLSLetsEncrypt.prototype['AcceptEULA'] = undefined;

/**
 * @member {String} Email
 */
InstallTLSLetsEncrypt.prototype['Email'] = undefined;

/**
 * @member {Boolean} StagingCA
 */
InstallTLSLetsEncrypt.prototype['StagingCA'] = undefined;

exports.default = InstallTLSLetsEncrypt;
