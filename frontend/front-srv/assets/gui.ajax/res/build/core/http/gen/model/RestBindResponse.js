/**
 * Pydio Cells Rest API
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * OpenAPI spec version: 1.0
 * 
 *
 * NOTE: This class is auto generated by the swagger code generator program.
 * https://github.com/swagger-api/swagger-codegen.git
 * Do not edit the class manually.
 *
 */

'use strict';

exports.__esModule = true;

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError('Cannot call a class as a function'); } }

var _ApiClient = require('../ApiClient');

var _ApiClient2 = _interopRequireDefault(_ApiClient);

/**
* The RestBindResponse model module.
* @module model/RestBindResponse
* @version 1.0
*/

var RestBindResponse = (function () {
    /**
    * Constructs a new <code>RestBindResponse</code>.
    * @alias module:model/RestBindResponse
    * @class
    */

    function RestBindResponse() {
        _classCallCheck(this, RestBindResponse);

        this.Success = undefined;
    }

    /**
    * Constructs a <code>RestBindResponse</code> from a plain JavaScript object, optionally creating a new instance.
    * Copies all relevant properties from <code>data</code> to <code>obj</code> if supplied or a new instance if not.
    * @param {Object} data The plain JavaScript object bearing properties of interest.
    * @param {module:model/RestBindResponse} obj Optional instance to populate.
    * @return {module:model/RestBindResponse} The populated <code>RestBindResponse</code> instance.
    */

    RestBindResponse.constructFromObject = function constructFromObject(data, obj) {
        if (data) {
            obj = obj || new RestBindResponse();

            if (data.hasOwnProperty('Success')) {
                obj['Success'] = _ApiClient2['default'].convertToType(data['Success'], 'Boolean');
            }
        }
        return obj;
    };

    /**
    * @member {Boolean} Success
    */
    return RestBindResponse;
})();

exports['default'] = RestBindResponse;
module.exports = exports['default'];
