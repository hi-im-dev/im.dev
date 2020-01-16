import React from 'react'
import { inject, observer } from 'mobx-react'
import { IntlProvider } from 'react-intl' /* react-intl imports */

import locale from '../../library/locale'
import {ISystem} from '../../store/system'

const Intl = inject('system')(observer((props:{system:ISystem} & any) =>{
    let {system} = props
    let messages = locale
    return (
        <>
            <IntlProvider locale={system.locale.split('_')[0]} messages={messages[system.locale.split('_')[0]]}>
                {props.children}
            </IntlProvider>
        </>
    )
}))

export default Intl