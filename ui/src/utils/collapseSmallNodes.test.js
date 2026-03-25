import { describe, expect, it } from 'vitest'
import collapseSmallNodes from './collapseSmallNodes'

describe('collapseSmallNodes', () => {
  it('collapses children below threshold into an (other) node', () => {
    const tree = {
      name: 'root',
      cost: 100,
      children: [
        { name: 'hot', cost: 80, children: [] },
        { name: 'tinyA', cost: 3, children: [] },
        { name: 'tinyB', cost: 2, children: [] },
      ],
    }

    expect(collapseSmallNodes(tree, 100, 0.05)).toEqual({
      name: 'root',
      cost: 100,
      children: [
        { name: 'hot', cost: 80, children: [] },
        {
          name: '(other)',
          cost: 5,
          children: [
            { name: 'tinyA', cost: 3, children: [] },
            { name: 'tinyB', cost: 2, children: [] },
          ],
        },
      ],
    })
  })
})
