import React from 'react';
// Import the original mapper
import MDXComponents from '@theme-original/MDXComponents';
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import AsciinemaWidget from '@site/src/components/AsciinemaWidget';

export default {
    // Re-use the default mapping
    ...MDXComponents,
    // Map the "highlight" tag to our <Highlight /> component!
    // `Highlight` will receive all props that were passed to `highlight` in MDX
    Tabs,
    TabItem,
    AsciinemaWidget,
};
